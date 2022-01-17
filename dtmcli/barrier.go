/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/go-redis/redis/v8"
)

// BarrierBusiFunc type for busi func
type BarrierBusiFunc func(tx *sql.Tx) error

// BranchBarrier every branch info
type BranchBarrier struct {
	TransType string
	Gid       string
	BranchID  string
	Op        string
	BarrierID int
}

func (bb *BranchBarrier) String() string {
	return fmt.Sprintf("transInfo: %s %s %s %s", bb.TransType, bb.Gid, bb.BranchID, bb.Op)
}

// BarrierFromQuery construct transaction info from request
func BarrierFromQuery(qs url.Values) (*BranchBarrier, error) {
	return BarrierFrom(qs.Get("trans_type"), qs.Get("gid"), qs.Get("branch_id"), qs.Get("op"))
}

// BarrierFrom construct transaction info from request
func BarrierFrom(transType, gid, branchID, op string) (*BranchBarrier, error) {
	ti := &BranchBarrier{
		TransType: transType,
		Gid:       gid,
		BranchID:  branchID,
		Op:        op,
	}
	if ti.TransType == "" || ti.Gid == "" || ti.BranchID == "" || ti.Op == "" {
		return nil, fmt.Errorf("invalid trans info: %v", ti)
	}
	return ti, nil
}

func insertBarrier(tx DB, transType string, gid string, branchID string, op string, barrierID string, reason string) (int64, error) {
	if op == "" {
		return 0, nil
	}
	sql := dtmimp.GetDBSpecial().GetInsertIgnoreTemplate(dtmimp.BarrierTableName+"(trans_type, gid, branch_id, op, barrier_id, reason) values(?,?,?,?,?,?)", "uniq_barrier")
	return dtmimp.DBExec(tx, sql, transType, gid, branchID, op, barrierID, reason)
}

// Call 子事务屏障，详细介绍见 https://zhuanlan.zhihu.com/p/388444465
// tx: 本地数据库的事务对象，允许子事务屏障进行事务操作
// busiCall: 业务函数，仅在必要时被调用
func (bb *BranchBarrier) Call(tx *sql.Tx, busiCall BarrierBusiFunc) (rerr error) {
	bb.BarrierID = bb.BarrierID + 1
	bid := fmt.Sprintf("%02d", bb.BarrierID)
	defer func() {
		// Logf("barrier call error is %v", rerr)
		if x := recover(); x != nil {
			_ = tx.Rollback()
			panic(x)
		} else if rerr != nil {
			_ = tx.Rollback()
		} else {
			rerr = tx.Commit()
		}
	}()
	ti := bb
	originOp := map[string]string{
		BranchCancel:     BranchTry,
		BranchCompensate: BranchAction,
	}[ti.Op]

	originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originOp, bid, ti.Op)
	currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.Op, bid, ti.Op)
	logger.Debugf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
	if (ti.Op == BranchCancel || ti.Op == BranchCompensate) && originAffected > 0 || // 这个是空补偿
		currentAffected == 0 { // 这个是重复请求或者悬挂
		return
	}
	rerr = busiCall(tx)
	return
}

// CallWithDB the same as Call, but with *sql.DB
func (bb *BranchBarrier) CallWithDB(db *sql.DB, busiCall BarrierBusiFunc) error {
	tx, err := db.Begin()
	if err == nil {
		err = bb.Call(tx, busiCall)
	}
	return err
}

// QueryPrepared queries prepared data
func (bb *BranchBarrier) QueryPrepared(db *sql.DB) error {
	_, err := insertBarrier(db, bb.TransType, bb.Gid, "00", "msg", "01", "rollback")
	var reason string
	if err == nil {
		sql := fmt.Sprintf("select reason from %s where gid=? and branch_id=? and op=? and barrier_id=?", dtmimp.BarrierTableName)
		err = db.QueryRow(sql, bb.Gid, "00", "msg", "01").Scan(&reason)
	}
	if reason == "rollback" {
		return ErrFailure
	}
	return err
}

// RedisCheckAdjustAmount check the value of key is valid and >= amount. then adjust the amount
func (bb *BranchBarrier) RedisCheckAdjustAmount(rd *redis.Client, key string, amount int, barrierExpire int) error {
	bkey1 := fmt.Sprintf("%s-%s-%s-%s-%02d", key, bb.Gid, bb.BranchID, bb.Op, bb.BarrierID)
	originOp := map[string]string{
		BranchCancel:     BranchTry,
		BranchCompensate: BranchAction,
	}[bb.Op]
	bkey2 := fmt.Sprintf("%s-%s-%s-%s-%02d", key, bb.Gid, bb.BranchID, originOp, bb.BarrierID)
	v, err := rd.Eval(rd.Context(), ` -- RedisCheckAdjustAmount
local v = redis.call('GET', KEYS[1])
local e1 = redis.call('GET', KEYS[2])

if v == false or v + ARGV[1] < 0 then
	return 'FAILURE'
end

if e1 ~= false then
	return
end

redis.call('SET', KEYS[2], 'op', 'EX', ARGV[3])

if ARGV[2] ~= '' then
	local e2 = redis.call('GET', KEYS[3])
	if e2 == false then
		redis.call('SET', KEYS[3], 'rollback', 'EX', ARGV[3])
		return
	end
end
redis.call('INCRBY', KEYS[1], ARGV[1])
`, []string{key, bkey1, bkey2}, amount, originOp, barrierExpire).Result()
	logger.Debugf("lua return v: %v err: %v", v, err)
	if err == redis.Nil {
		err = nil
	}
	if err == nil && v == ResultFailure {
		err = ErrFailure
	}
	return err
}

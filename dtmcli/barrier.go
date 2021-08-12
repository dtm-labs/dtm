package dtmcli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
)

// BusiFunc type for busi func
type BusiFunc func(db *sql.Tx) (interface{}, error)

// BranchBarrier every branch info
type BranchBarrier struct {
	TransType  string
	Gid        string
	BranchID   string
	BranchType string
	BarrierID  int
}

func (bb *BranchBarrier) String() string {
	return fmt.Sprintf("transInfo: %s %s %s %s", bb.TransType, bb.Gid, bb.BranchID, bb.BranchType)
}

// BarrierFromQuery construct transaction info from request
func BarrierFromQuery(qs url.Values) (*BranchBarrier, error) {
	return BarrierFrom(qs.Get("trans_type"), qs.Get("gid"), qs.Get("branch_id"), qs.Get("branch_type"))
}

// BarrierFrom construct transaction info from request
func BarrierFrom(transType, gid, branchID, branchType string) (*BranchBarrier, error) {
	ti := &BranchBarrier{
		TransType:  transType,
		Gid:        gid,
		BranchID:   branchID,
		BranchType: branchType,
	}
	if ti.TransType == "" || ti.Gid == "" || ti.BranchID == "" || ti.BranchType == "" {
		return nil, fmt.Errorf("invlid trans info: %v", ti)
	}
	return ti, nil
}

func insertBarrier(tx *sql.Tx, transType string, gid string, branchID string, branchType string, barrierID string, reason string) (int64, error) {
	if branchType == "" {
		return 0, nil
	}
	return StxExec(tx, "insert ignore into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type, barrier_id, reason) values(?,?,?,?,?,?)", transType, gid, branchID, branchType, barrierID, reason)
}

// Call 子事务屏障，详细介绍见 https://zhuanlan.zhihu.com/p/388444465
// db: 本地数据库
// transInfo: 事务信息
// bisiCall: 业务函数，仅在必要时被调用
// 返回值:
// 如果正常调用，返回bisiCall的结果
// 如果发生重复调用，则busiCall不会被重复调用，直接对保存在数据库中上一次的结果，进行unmarshal，通常是一个map[string]interface{}，直接作为http的resp
// 如果发生悬挂，则busiCall不会被调用，直接返回错误 {"dtm_result": "FAILURE"}
// 如果发生空补偿，则busiCall不会被调用，直接返回 {"dtm_result": "SUCCESS"}
func (bb *BranchBarrier) Call(db *sql.DB, busiCall BusiFunc) (res interface{}, rerr error) {
	bb.BarrierID = bb.BarrierID + 1
	bid := fmt.Sprintf("%02d", bb.BarrierID)
	tx, rerr := db.BeginTx(context.Background(), &sql.TxOptions{})
	if rerr != nil {
		return
	}
	defer func() {
		Logf("result is %v error is %v", res, rerr)
		if x := recover(); x != nil {
			tx.Rollback()
			panic(x)
		} else if rerr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	ti := bb
	originType := map[string]string{
		"cancel":     "try",
		"compensate": "action",
	}[ti.BranchType]
	originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.BranchType)
	currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.BranchType, bid, ti.BranchType)
	Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
	if (ti.BranchType == "cancel" || ti.BranchType == "compensate") && originAffected > 0 { // 这个是空补偿，返回成功
		res = ResultSuccess
		return
	} else if currentAffected == 0 { // 插入不成功
		var result sql.NullString
		err := StxQueryRow(tx, "select result from dtm_barrier.barrier where trans_type=? and gid=? and branch_id=? and branch_type=? and barrier_id=? and reason=?",
			ti.TransType, ti.Gid, ti.BranchID, ti.BranchType, bid, ti.BranchType).Scan(&result)
		if err == sql.ErrNoRows { // 这个是悬挂操作，返回失败，AP收到这个返回，会尽快回滚
			res = ResultFailure
			return
		}
		if err != nil {
			rerr = err
			return
		}
		if result.Valid { // 数据库里有上一次结果，返回上一次的结果
			rerr = json.Unmarshal([]byte(result.String), &res)
			return
		}
		// 数据库里没有上次的结果，属于重复空补偿，直接返回成功
		res = ResultSuccess
		return
	}
	res, rerr = busiCall(tx)
	if rerr == nil { // 正确返回了，需要将结果保存到数据库
		sval := MustMarshalString(res)
		_, rerr = StxExec(tx, "update dtm_barrier.barrier set result=? where trans_type=? and gid=? and branch_id=? and branch_type=?", sval,
			ti.TransType, ti.Gid, ti.BranchID, ti.BranchType)
	}
	return
}

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
)

// BarrierBusiFunc type for busi func
type BarrierBusiFunc func(tx *sql.Tx) error

// BranchBarrier every branch info
type BranchBarrier struct {
	TransType        string
	Gid              string
	BranchID         string
	Op               string
	BarrierID        int
	DBType           string // DBTypeMysql | DBTypePostgres
	BarrierTableName string
}

func (bb *BranchBarrier) String() string {
	return fmt.Sprintf("transInfo: %s %s %s %s", bb.TransType, bb.Gid, bb.BranchID, bb.Op)
}

func (bb *BranchBarrier) newBarrierID() string {
	bb.BarrierID++
	return fmt.Sprintf("%02d", bb.BarrierID)
}

// BarrierFromQuery construct transaction info from request
func BarrierFromQuery(qs url.Values) (*BranchBarrier, error) {
	return BarrierFrom(dtmimp.EscapeGet(qs, "trans_type"), dtmimp.EscapeGet(qs, "gid"), dtmimp.EscapeGet(qs, "branch_id"), dtmimp.EscapeGet(qs, "op"))
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

// Call see detail description in https://en.dtm.pub/practice/barrier.html
// tx: local transaction connection
// busiCall: busi func
func (bb *BranchBarrier) Call(tx *sql.Tx, busiCall BarrierBusiFunc) (rerr error) {
	bid := bb.newBarrierID()
	defer dtmimp.DeferDo(&rerr, func() error {
		return tx.Commit()
	}, func() error {
		return tx.Rollback()
	})
	originOp := map[string]string{
		dtmimp.OpCancel:     dtmimp.OpTry,
		dtmimp.OpCompensate: dtmimp.OpAction,
	}[bb.Op]

	originAffected, oerr := dtmimp.InsertBarrier(tx, bb.TransType, bb.Gid, bb.BranchID, originOp, bid, bb.Op, bb.DBType, bb.BarrierTableName)
	currentAffected, rerr := dtmimp.InsertBarrier(tx, bb.TransType, bb.Gid, bb.BranchID, bb.Op, bid, bb.Op, bb.DBType, bb.BarrierTableName)
	logger.Debugf("originAffected: %d currentAffected: %d", originAffected, currentAffected)

	if rerr == nil && bb.Op == dtmimp.MsgDoOp && currentAffected == 0 { // for msg's DoAndSubmit, repeated insert should be rejected.
		return ErrDuplicated
	}

	if rerr == nil {
		rerr = oerr
	}

	if (bb.Op == dtmimp.OpCancel || bb.Op == dtmimp.OpCompensate) && originAffected > 0 || // null compensate
		currentAffected == 0 { // repeated request or dangled request
		return
	}
	if rerr == nil {
		rerr = busiCall(tx)
	}
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
	_, err := dtmimp.InsertBarrier(db, bb.TransType, bb.Gid, dtmimp.MsgDoBranch0, dtmimp.MsgDoOp, dtmimp.MsgDoBarrier1, dtmimp.OpRollback, bb.DBType, bb.BarrierTableName)
	var reason string
	if err == nil {
		sql := fmt.Sprintf("select reason from %s where gid=? and branch_id=? and op=? and barrier_id=?", dtmimp.BarrierTableName)
		sql = dtmimp.GetDBSpecial(bb.DBType).GetPlaceHoldSQL(sql)
		logger.Debugf("queryrow: %s", sql, bb.Gid, dtmimp.MsgDoBranch0, dtmimp.MsgDoOp, dtmimp.MsgDoBarrier1)
		err = db.QueryRow(sql, bb.Gid, dtmimp.MsgDoBranch0, dtmimp.MsgDoOp, dtmimp.MsgDoBarrier1).Scan(&reason)
	}
	if reason == dtmimp.OpRollback {
		return ErrFailure
	}
	return err
}

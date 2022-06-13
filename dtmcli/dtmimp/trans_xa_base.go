/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"database/sql"
	"strings"
)

// XaHandlePhase2 Handle the callback of commit/rollback
func XaHandlePhase2(gid string, dbConf DBConf, branchID string, op string) error {
	db, err := PooledDB(dbConf)
	if err != nil {
		return err
	}
	xaID := gid + "-" + branchID
	_, err = DBExec(dbConf.Driver, db, GetDBSpecial(dbConf.Driver).GetXaSQL(op, xaID))
	if err != nil &&
		(strings.Contains(err.Error(), "XAER_NOTA") || strings.Contains(err.Error(), "does not exist")) { // Repeat commit/rollback with the same id, report this error, ignore
		err = nil
	}
	if op == OpRollback && err == nil {
		// rollback insert a row after prepare. no-error means prepare has finished.
		_, err = InsertBarrier(db, "xa", gid, branchID, OpAction, XaBarrier1, op, dbConf.Driver, "")
	}
	return err
}

// XaHandleLocalTrans public handler of LocalTransaction via http/grpc
func XaHandleLocalTrans(xa *TransBase, dbConf DBConf, cb func(*sql.DB) error) (rerr error) {
	xaBranch := xa.Gid + "-" + xa.BranchID
	db, rerr := XaDB(dbConf)
	if rerr != nil {
		return
	}
	defer func() { _ = db.Close() }()
	defer DeferDo(&rerr, func() error {
		_, err := DBExec(dbConf.Driver, db, GetDBSpecial(dbConf.Driver).GetXaSQL("prepare", xaBranch))
		return err
	}, func() error {
		return nil
	})
	_, rerr = DBExec(dbConf.Driver, db, GetDBSpecial(dbConf.Driver).GetXaSQL("start", xaBranch))
	if rerr != nil {
		return
	}
	defer func() {
		_, _ = DBExec(dbConf.Driver, db, GetDBSpecial(dbConf.Driver).GetXaSQL("end", xaBranch))
	}()
	// prepare and rollback both insert a row
	_, rerr = InsertBarrier(db, xa.TransType, xa.Gid, xa.BranchID, OpAction, XaBarrier1, OpAction, dbConf.Driver, "")
	if rerr == nil {
		rerr = cb(db)
	}
	return
}

// XaHandleGlobalTrans http/grpc GlobalTransaction shared func
func XaHandleGlobalTrans(xa *TransBase, callDtm func(string) error, callBusi func() error) (rerr error) {
	rerr = callDtm("prepare")
	if rerr != nil {
		return
	}
	defer DeferDo(&rerr, func() error {
		return callDtm("submit")
	}, func() error {
		return callDtm("abort")
	})
	rerr = callBusi()
	return
}

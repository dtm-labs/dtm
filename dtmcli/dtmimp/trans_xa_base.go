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

// XaClientBase XaClient/XaGrpcClient base. shared by http and grpc
type XaClientBase struct {
	Server    string
	Conf      DBConf
	NotifyURL string
}

// HandleCallback Handle the callback of commit/rollback
func (xc *XaClientBase) HandleCallback(gid string, branchID string, action string) error {
	db, err := PooledDB(xc.Conf)
	if err != nil {
		return err
	}
	xaID := gid + "-" + branchID
	_, err = DBExec(db, GetDBSpecial().GetXaSQL(action, xaID))
	if err != nil &&
		(strings.Contains(err.Error(), "XAER_NOTA") || strings.Contains(err.Error(), "does not exist")) { // Repeat commit/rollback with the same id, report this error, ignore
		err = nil
	}
	if action == OpRollback && err == nil {
		// rollback insert a row after prepare. no-error means prepare has finished.
		_, err = InsertBarrier(db, "xa", gid, branchID, OpAction, XaBarrier1, action)
	}
	return err
}

// HandleLocalTrans public handler of LocalTransaction via http/grpc
func (xc *XaClientBase) HandleLocalTrans(xa *TransBase, cb func(*sql.DB) error) (rerr error) {
	xaBranch := xa.Gid + "-" + xa.BranchID
	db, rerr := StandaloneDB(xc.Conf)
	if rerr != nil {
		return
	}
	defer func() { _ = db.Close() }()
	defer DeferDo(&rerr, func() error {
		_, err := DBExec(db, GetDBSpecial().GetXaSQL("prepare", xaBranch))
		return err
	}, func() error {
		return nil
	})
	_, rerr = DBExec(db, GetDBSpecial().GetXaSQL("start", xaBranch))
	if rerr != nil {
		return
	}
	defer func() {
		_, _ = DBExec(db, GetDBSpecial().GetXaSQL("end", xaBranch))
	}()
	// prepare and rollback both insert a row
	_, rerr = InsertBarrier(db, xa.TransType, xa.Gid, xa.BranchID, OpAction, XaBarrier1, OpAction)
	if rerr == nil {
		rerr = cb(db)
	}
	return
}

// HandleGlobalTrans http/grpc GlobalTransaction shared func
func (xc *XaClientBase) HandleGlobalTrans(xa *TransBase, callDtm func(string) error, callBusi func() error) (rerr error) {
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

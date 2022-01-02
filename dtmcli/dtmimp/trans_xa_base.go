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

// HandleCallback type of commit/rollback callback handler
func (xc *XaClientBase) HandleCallback(gid string, branchID string, action string) error {
	db, err := StandaloneDB(xc.Conf)
	if err != nil {
		return err
	}
	defer db.Close()
	xaID := gid + "-" + branchID
	_, err = DBExec(db, GetDBSpecial().GetXaSQL(action, xaID))
	if err != nil &&
		(strings.Contains(err.Error(), "XAER_NOTA") || strings.Contains(err.Error(), "does not exist")) { // Repeat commit/rollback with the same id, report this error, ignore
		err = nil
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
	defer func() { db.Close() }()
	defer func() {
		x := recover()
		_, err := DBExec(db, GetDBSpecial().GetXaSQL("end", xaBranch))
		if x == nil && rerr == nil && err == nil {
			_, err = DBExec(db, GetDBSpecial().GetXaSQL("prepare", xaBranch))
		}
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = DBExec(db, GetDBSpecial().GetXaSQL("start", xaBranch))
	if rerr != nil {
		return
	}
	rerr = cb(db)
	return
}

// HandleGlobalTrans http/grpc GlobalTransaction的公共方法
func (xc *XaClientBase) HandleGlobalTrans(xa *TransBase, callDtm func(string) error, callBusi func() error) (rerr error) {
	rerr = callDtm("prepare")
	if rerr != nil {
		return
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := If(x != nil || rerr != nil, "abort", "submit").(string)
		err := callDtm(operation)
		if rerr == nil { // 如果用户函数没有返回错误，那么返回dtm的
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	rerr = callBusi()
	return
}

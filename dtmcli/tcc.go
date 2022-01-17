/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"fmt"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/go-resty/resty/v2"
)

// Tcc struct of tcc
type Tcc struct {
	dtmimp.TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *Tcc) (*resty.Response, error)

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm server address
// gid global transaction ID
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	return TccGlobalTransaction2(dtm, gid, func(t *Tcc) {}, tccFunc)
}

// TccGlobalTransaction2 new version of TccGlobalTransaction, add custom param
func TccGlobalTransaction2(dtm string, gid string, custom func(*Tcc), tccFunc TccGlobalFunc) (rerr error) {
	tcc := &Tcc{TransBase: *dtmimp.NewTransBase(gid, "tcc", dtm, "")}
	custom(tcc)
	rerr = dtmimp.TransCallDtm(&tcc.TransBase, tcc, "prepare")
	if rerr != nil {
		return rerr
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := dtmimp.If(x == nil && rerr == nil, "submit", "abort").(string)
		err := dtmimp.TransCallDtm(&tcc.TransBase, tcc, operation)
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = tccFunc(tcc)
	return
}

// TccFromQuery tcc from request info
func TccFromQuery(qs url.Values) (*Tcc, error) {
	tcc := &Tcc{TransBase: *dtmimp.TransBaseFromQuery(qs)}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s parentID: %s", tcc.Dtm, tcc.Gid, tcc.BranchID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
func (t *Tcc) CallBranch(body interface{}, tryURL string, confirmURL string, cancelURL string) (*resty.Response, error) {
	branchID := t.NewSubBranchID()
	err := dtmimp.TransRegisterBranch(&t.TransBase, map[string]string{
		"data":        dtmimp.MustMarshalString(body),
		"branch_id":   branchID,
		BranchConfirm: confirmURL,
		BranchCancel:  cancelURL,
	}, "registerBranch")
	if err != nil {
		return nil, err
	}
	return dtmimp.TransRequestBranch(&t.TransBase, body, branchID, BranchTry, tryURL)
}

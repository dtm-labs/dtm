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
// tccFunc define the detail tcc busi
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
	defer dtmimp.DeferDo(&rerr, func() error {
		return dtmimp.TransCallDtm(&tcc.TransBase, tcc, "submit")
	}, func() error {
		if rerr != nil {
			tcc.RollbackReason = rerr.Error()
		}
		return dtmimp.TransCallDtm(&tcc.TransBase, tcc, "abort")
	})
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
		"data":           dtmimp.MustMarshalString(body),
		"branch_id":      branchID,
		dtmimp.OpConfirm: confirmURL,
		dtmimp.OpCancel:  cancelURL,
	}, "registerBranch")
	if err != nil {
		return nil, err
	}
	return dtmimp.TransRequestBranch(&t.TransBase, "POST", body, branchID, dtmimp.OpTry, tryURL)
}

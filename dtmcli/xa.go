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
	"github.com/go-resty/resty/v2"
)

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *Xa) (*resty.Response, error)

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *Xa) error

// Xa xa transaction
type Xa struct {
	dtmimp.TransBase
	Phase2URL string
}

// XaFromQuery construct xa info from request
func XaFromQuery(qs url.Values) (*Xa, error) {
	xa := &Xa{TransBase: *dtmimp.TransBaseFromQuery(qs)}
	xa.Op = dtmimp.EscapeGet(qs, "op")
	xa.Phase2URL = dtmimp.EscapeGet(qs, "phase2_url")
	if xa.Gid == "" || xa.BranchID == "" || xa.Op == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s branchid: %s op: %s phase2_url: %s", xa.Gid, xa.BranchID, xa.Op, xa.Phase2URL)
	}
	return xa, nil
}

// XaLocalTransaction start a xa local transaction
func XaLocalTransaction(qs url.Values, dbConf DBConf, xaFunc XaLocalFunc) error {
	xa, err := XaFromQuery(qs)
	if err != nil {
		return err
	}
	if xa.Op == dtmimp.OpCommit || xa.Op == dtmimp.OpRollback {
		return dtmimp.XaHandlePhase2(xa.Gid, dbConf, xa.BranchID, xa.Op)
	}
	return dtmimp.XaHandleLocalTrans(&xa.TransBase, dbConf, func(db *sql.DB) error {
		err := xaFunc(db, xa)
		if err != nil {
			return err
		}
		return dtmimp.TransRegisterBranch(&xa.TransBase, map[string]string{
			"url":       xa.Phase2URL,
			"branch_id": xa.BranchID,
		}, "registerBranch")
	})
}

// XaGlobalTransaction start a xa global transaction
func XaGlobalTransaction(server string, gid string, xaFunc XaGlobalFunc) error {
	return XaGlobalTransaction2(server, gid, func(x *Xa) {}, xaFunc)
}

// XaGlobalTransaction2 start a xa global transaction with xa custom function
func XaGlobalTransaction2(server string, gid string, custom func(*Xa), xaFunc XaGlobalFunc) (rerr error) {
	xa := &Xa{TransBase: *dtmimp.NewTransBase(gid, "xa", server, "")}
	custom(xa)
	return dtmimp.XaHandleGlobalTrans(&xa.TransBase, func(action string) error {
		return dtmimp.TransCallDtm(&xa.TransBase, action)
	}, func() error {
		_, rerr := xaFunc(xa)
		return rerr
	})
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewSubBranchID()
	return dtmimp.TransRequestBranch(&x.TransBase, "POST", body, branchID, dtmimp.OpAction, url)
}

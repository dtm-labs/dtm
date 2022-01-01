/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

// BranchIDGen used to generate a sub branch id
type BranchIDGen struct {
	BranchID    string
	subBranchID int
}

// NewSubBranchID generate a sub branch id
func (g *BranchIDGen) NewSubBranchID() string {
	if g.subBranchID >= 99 {
		panic(fmt.Errorf("branch id is larger than 99"))
	}
	if len(g.BranchID) >= 20 {
		panic(fmt.Errorf("total branch id is longer than 20"))
	}
	g.subBranchID = g.subBranchID + 1
	return g.CurrentSubBranchID()
}

// CurrentSubBranchID return current branchID
func (g *BranchIDGen) CurrentSubBranchID() string {
	return g.BranchID + fmt.Sprintf("%02d", g.subBranchID)
}

// TransOptions transaction options
type TransOptions struct {
	WaitResult         bool              `json:"wait_result,omitempty" gorm:"-"`
	TimeoutToFail      int64             `json:"timeout_to_fail,omitempty" gorm:"-"` // for trans type: xa, tcc
	RetryInterval      int64             `json:"retry_interval,omitempty" gorm:"-"`  // for trans type: msg saga xa tcc
	PassthroughHeaders []string          `json:"passthrough_headers,omitempty" gorm:"-"`
	BranchHeaders      map[string]string `json:"branch_headers,omitempty" gorm:"-"`
}

// TransBase base for all trans
type TransBase struct {
	Gid        string `json:"gid"`
	TransType  string `json:"trans_type"`
	Dtm        string `json:"-"`
	CustomData string `json:"custom_data,omitempty"`
	TransOptions

	Steps       []map[string]string `json:"steps,omitempty"`    // use in MSG/SAGA
	Payloads    []string            `json:"payloads,omitempty"` // used in MSG/SAGA
	BinPayloads [][]byte            `json:"-"`
	BranchIDGen `json:"-"`          // used in XA/TCC
	Op          string              `json:"-"` // used in XA/TCC

	QueryPrepared string `json:"query_prepared,omitempty"` // used in MSG
}

// NewTransBase new a TransBase
func NewTransBase(gid string, transType string, dtm string, branchID string) *TransBase {
	return &TransBase{
		Gid:          gid,
		TransType:    transType,
		BranchIDGen:  BranchIDGen{BranchID: branchID},
		Dtm:          dtm,
		TransOptions: TransOptions{PassthroughHeaders: PassthroughHeaders},
	}
}

// TransBaseFromQuery construct transaction info from request
func TransBaseFromQuery(qs url.Values) *TransBase {
	return NewTransBase(qs.Get("gid"), qs.Get("trans_type"), qs.Get("dtm"), qs.Get("branch_id"))
}

// TransCallDtm TransBase call dtm
func TransCallDtm(tb *TransBase, body interface{}, operation string) error {
	resp, err := RestyClient.R().
		SetBody(body).Post(fmt.Sprintf("%s/%s", tb.Dtm, operation))
	if err != nil {
		return err
	}
	if !strings.Contains(resp.String(), ResultSuccess) {
		return errors.New(resp.String())
	}
	return nil
}

// TransRegisterBranch TransBase register a branch to dtm
func TransRegisterBranch(tb *TransBase, added map[string]string, operation string) error {
	m := map[string]string{
		"gid":        tb.Gid,
		"trans_type": tb.TransType,
	}
	for k, v := range added {
		m[k] = v
	}
	return TransCallDtm(tb, m, operation)
}

// TransRequestBranch TransBAse request branch result
func TransRequestBranch(t *TransBase, body interface{}, branchID string, op string, url string) (*resty.Response, error) {
	resp, err := RestyClient.R().
		SetBody(body).
		SetQueryParams(map[string]string{
			"dtm":        t.Dtm,
			"gid":        t.Gid,
			"branch_id":  branchID,
			"trans_type": t.TransType,
			"op":         op,
		}).
		SetHeaders(t.BranchHeaders).
		Post(url)
	return resp, CheckResponse(resp, err)
}

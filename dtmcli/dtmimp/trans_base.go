/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	TimeoutToFail      int64             `json:"timeout_to_fail,omitempty" gorm:"-"`     // for trans type: xa, tcc, unit: second
	RequestTimeout     int64             `json:"request_timeout,omitempty" gorm:"-"`     // for global trans resets request timeout, unit: second
	RetryInterval      int64             `json:"retry_interval,omitempty" gorm:"-"`      // for trans type: msg saga xa tcc, unit: second
	PassthroughHeaders []string          `json:"passthrough_headers,omitempty" gorm:"-"` // for inherit the specified gin context headers
	BranchHeaders      map[string]string `json:"branch_headers,omitempty" gorm:"-"`      // custom branch headers,  dtm server => service api
	Concurrent         bool              `json:"concurrent" gorm:"-"`                    // for trans type: saga msg
	RollbackReason     string            `json:"rollback_reason,omitempty" gorm:"-"`
}

// TransBase base for all trans
type TransBase struct {
	Gid        string `json:"gid"` //  NOTE: unique in storage, can customize the generation rules instead of using server-side generation, it will help with the tracking
	TransType  string `json:"trans_type"`
	Dtm        string `json:"-"`
	CustomData string `json:"custom_data,omitempty"` // nosql data persistence
	TransOptions
	Context context.Context `json:"-" gorm:"-"`

	Steps       []map[string]string `json:"steps,omitempty"`    // use in MSG/SAGA
	Payloads    []string            `json:"payloads,omitempty"` // used in MSG/SAGA
	BinPayloads [][]byte            `json:"-"`
	BranchIDGen `json:"-"`          // used in XA/TCC
	Op          string              `json:"-"` // used in XA/TCC

	QueryPrepared string `json:"query_prepared,omitempty"` // used in MSG
	Protocol      string `json:"protocol"`
}

// NewTransBase new a TransBase
func NewTransBase(gid string, transType string, dtm string, branchID string) *TransBase {
	return &TransBase{
		Gid:          gid,
		TransType:    transType,
		BranchIDGen:  BranchIDGen{BranchID: branchID},
		Dtm:          dtm,
		TransOptions: TransOptions{PassthroughHeaders: PassthroughHeaders},
		Context:      context.Background(),
	}
}

// WithGlobalTransRequestTimeout defines global trans request timeout
func (t *TransBase) WithGlobalTransRequestTimeout(timeout int64) {
	t.RequestTimeout = timeout
}

// TransBaseFromQuery construct transaction info from request
func TransBaseFromQuery(qs url.Values) *TransBase {
	return NewTransBase(EscapeGet(qs, "gid"), EscapeGet(qs, "trans_type"), EscapeGet(qs, "dtm"), EscapeGet(qs, "branch_id"))
}

// TransCallDtmExt TransBase call dtm
func TransCallDtmExt(tb *TransBase, body interface{}, operation string) (*resty.Response, error) {
	if tb.Protocol == Jrpc {
		return transCallDtmJrpc(tb, body, operation)
	}
	if tb.RequestTimeout != 0 {
		RestyClient.SetTimeout(time.Duration(tb.RequestTimeout) * time.Second)
	}
	resp, err := RestyClient.R().
		SetBody(body).Post(fmt.Sprintf("%s/%s", tb.Dtm, operation))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || strings.Contains(resp.String(), ResultFailure) {
		return nil, errors.New(resp.String())
	}
	return resp, nil
}

func TransCallDtm(tb *TransBase, operation string) error {
	_, err := TransCallDtmExt(tb, tb, operation)
	return err
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
	_, err := TransCallDtmExt(tb, m, operation)
	return err
}

// TransRequestBranch TransBase request branch result
func TransRequestBranch(t *TransBase, method string, body interface{}, branchID string, op string, url string) (*resty.Response, error) {
	if url == "" {
		return nil, nil
	}
	query := map[string]string{
		"dtm":        t.Dtm,
		"gid":        t.Gid,
		"branch_id":  branchID,
		"trans_type": t.TransType,
		"op":         op,
	}
	if t.TransType == "xa" { // xa trans will add notify_url
		query["phase2_url"] = url
	}
	resp, err := RestyClient.R().
		SetBody(body).
		SetQueryParams(query).
		SetHeaders(t.BranchHeaders).
		Execute(method, url)
	if err == nil {
		err = RespAsErrorCompatible(resp)
	}
	return resp, err
}

func transCallDtmJrpc(tb *TransBase, body interface{}, operation string) (*resty.Response, error) {
	if tb.RequestTimeout != 0 {
		RestyClient.SetTimeout(time.Duration(tb.RequestTimeout) * time.Second)
	}
	var result map[string]interface{}
	resp, err := RestyClient.R().
		SetBody(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "no-use",
			"method":  operation,
			"params":  body,
		}).
		SetResult(&result).
		Post(tb.Dtm)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || result["error"] != nil {
		return nil, errors.New(resp.String())
	}
	return resp, nil
}

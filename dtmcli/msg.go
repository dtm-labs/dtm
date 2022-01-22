/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"database/sql"
	"errors"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
)

// Msg reliable msg type
type Msg struct {
	dtmimp.TransBase
}

// NewMsg create new msg
func NewMsg(server string, gid string) *Msg {
	return &Msg{TransBase: *dtmimp.NewTransBase(gid, "msg", server, "")}
}

// Add add a new step
func (s *Msg) Add(action string, postData interface{}) *Msg {
	s.Steps = append(s.Steps, map[string]string{"action": action})
	s.Payloads = append(s.Payloads, dtmimp.MustMarshalString(postData))
	return s
}

// Prepare prepare the msg, msg will later be submitted
func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = dtmimp.OrString(queryPrepared, s.QueryPrepared)
	return dtmimp.TransCallDtm(&s.TransBase, s, "prepare")
}

// Submit submit the msg
func (s *Msg) Submit() error {
	return dtmimp.TransCallDtm(&s.TransBase, s, "submit")
}

// PrepareAndSubmit one method for the entire busi->prepare->submit
func (s *Msg) PrepareAndSubmit(queryPrepared string, db *sql.DB, busiCall BarrierBusiFunc) error {
	return s.PrepareAndSubmitBarrier(queryPrepared, func(bb *BranchBarrier) error {
		return bb.CallWithDB(db, busiCall)
	})
}

// PrepareAndSubmitBarrier one method for the entire busi->prepare->submit
func (s *Msg) PrepareAndSubmitBarrier(queryPrepared string, busiCall func(bb *BranchBarrier) error) error {
	bb, err := BarrierFrom(s.TransType, s.Gid, "00", "msg") // a special barrier for msg QueryPrepared
	if err == nil {
		err = s.Prepare(queryPrepared)
	}
	if err == nil {
		err = busiCall(bb)
		if err != nil && !errors.Is(err, ErrFailure) {
			_, err = dtmimp.TransRequestBranch(&s.TransBase, "GET", nil, bb.BranchID, bb.Op, queryPrepared)
		}
		if errors.Is(err, ErrFailure) {
			_ = dtmimp.TransCallDtm(&s.TransBase, s, "abort")
		}
	}
	if err == nil {
		err = s.Submit()
	}
	return err
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"database/sql"

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

// PrepareAndSubmit execs prepare and submit operation
func (s *Msg) PrepareAndSubmit(queryPrepared string, db *sql.DB, busiCall BarrierBusiFunc) error {
	bb, err := BarrierFrom(s.TransType, s.Gid, "00", "msg") // a special barrier for msg QueryPrepared
	if err == nil {
		err = bb.CallWithDB(db, func(tx *sql.Tx) error {
			err := busiCall(tx)
			if err == nil {
				err = s.Prepare(queryPrepared)
			}
			return err
		})
	}
	if err == nil {
		err = s.Submit()
	}
	return err
}

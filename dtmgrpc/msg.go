/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"database/sql"
	"errors"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/protobuf/proto"
)

// MsgGrpc reliable msg type
type MsgGrpc struct {
	dtmcli.Msg
}

// NewMsgGrpc create new msg
func NewMsgGrpc(server string, gid string) *MsgGrpc {
	return &MsgGrpc{Msg: *dtmcli.NewMsg(server, gid)}
}

// Add add a new step
func (s *MsgGrpc) Add(action string, msg proto.Message) *MsgGrpc {
	s.Steps = append(s.Steps, map[string]string{"action": action})
	s.BinPayloads = append(s.BinPayloads, dtmgimp.MustProtoMarshal(msg))
	return s
}

// Prepare prepare the msg, msg will later be submitted
func (s *MsgGrpc) Prepare(queryPrepared string) error {
	s.QueryPrepared = dtmimp.OrString(queryPrepared, s.QueryPrepared)
	return dtmgimp.DtmGrpcCall(&s.TransBase, "Prepare")
}

// Submit submit the msg
func (s *MsgGrpc) Submit() error {
	return dtmgimp.DtmGrpcCall(&s.TransBase, "Submit")
}

// DoAndSubmitDB short method for Do on db type. please see DoAndSubmit
func (s *MsgGrpc) DoAndSubmitDB(queryPrepared string, db *sql.DB, busiCall dtmcli.BarrierBusiFunc) error {
	return s.DoAndSubmit(queryPrepared, func(bb *dtmcli.BranchBarrier) error {
		return bb.CallWithDB(db, busiCall)
	})
}

// DoAndSubmit one method for the entire prepare->busi->submit
// the error returned by busiCall will be returned
// if busiCall return ErrFailure, then abort is called directly
// if busiCall return not nil error other than ErrFailure, then DoAndSubmit will call queryPrepared to get the result
func (s *MsgGrpc) DoAndSubmit(queryPrepared string, busiCall func(bb *dtmcli.BranchBarrier) error) error {
	bb, err := dtmcli.BarrierFrom(s.TransType, s.Gid, "00", "msg") // a special barrier for msg QueryPrepared
	if err == nil {
		err = s.Prepare(queryPrepared)
	}
	if err == nil {
		errb := busiCall(bb)
		if errb != nil && !errors.Is(errb, dtmcli.ErrFailure) {
			err = dtmgimp.InvokeBranch(&s.TransBase, true, nil, queryPrepared, &[]byte{}, bb.BranchID, bb.Op)
			err = GrpcError2DtmError(err)
		}
		if errors.Is(errb, dtmcli.ErrFailure) || errors.Is(err, dtmcli.ErrFailure) {
			_ = dtmgimp.DtmGrpcCall(&s.TransBase, "Abort")
		} else if err == nil {
			err = s.Submit()
		}
		if errb != nil {
			return errb
		}
	}
	return err
}

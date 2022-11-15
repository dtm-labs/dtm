/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// MsgGrpc reliable msg type
type MsgGrpc struct {
	dtmcli.Msg
}

// NewMsgGrpc create new msg
func NewMsgGrpc(server string, gid string, opts ...TransBaseOption) *MsgGrpc {
	mg := &MsgGrpc{Msg: *dtmcli.NewMsg(server, gid)}

	for _, opt := range opts {
		opt(&mg.TransBase)
	}

	return mg
}

// Add add a new step
func (s *MsgGrpc) Add(action string, msg proto.Message) *MsgGrpc {
	s.Steps = append(s.Steps, map[string]string{"action": action})
	s.BinPayloads = append(s.BinPayloads, dtmgimp.MustProtoMarshal(msg))
	return s
}

// AddTopic add a new topic step
func (s *MsgGrpc) AddTopic(topic string, msg proto.Message) *MsgGrpc {
	return s.Add(fmt.Sprintf("%s%s", dtmimp.MsgTopicPrefix, topic), msg)
}

// SetDelay delay call branch, unit second
func (s *MsgGrpc) SetDelay(delay uint64) *MsgGrpc {
	s.Msg.SetDelay(delay)
	return s
}

// Prepare prepare the msg, msg will later be submitted
func (s *MsgGrpc) Prepare(queryPrepared string) error {
	s.QueryPrepared = dtmimp.OrString(queryPrepared, s.QueryPrepared)
	return dtmgimp.DtmGrpcCall(&s.TransBase, "Prepare")
}

// Submit submit the msg
func (s *MsgGrpc) Submit() error {
	s.Msg.BuildCustomOptions()
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
func (s *MsgGrpc) DoAndSubmit(queryPrepared string, busiCall func(bb *dtmcli.BranchBarrier) error, opts ...grpc.CallOption) error {
	bb, err := dtmcli.BarrierFrom(s.TransType, s.Gid, dtmimp.MsgDoBranch0, dtmimp.MsgDoOp) // a special barrier for msg QueryPrepared
	if err == nil {
		err = s.Prepare(queryPrepared)
	}
	if err == nil {
		errb := busiCall(bb)
		if errb != nil && !errors.Is(errb, dtmcli.ErrFailure) {
			err = dtmgimp.InvokeBranch(&s.TransBase, true, nil, queryPrepared, &[]byte{}, bb.BranchID, bb.Op, opts...)
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

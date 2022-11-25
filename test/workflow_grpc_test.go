/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"database/sql"
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestWorkflowGrpcSimple(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	req := &busi.ReqGrpc{Amount: 30, TransInResult: "FAILURE"}
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqGrpc
		dtmgimp.MustProtoUnmarshal(data, &req)
		_, err := busi.BusiCli.TransOutBSaga(wf.NewBranchCtx(), &req)
		if err != nil {
			return err
		}
		_, err = busi.BusiCli.TransInBSaga(wf.NewBranchCtx(), &req)
		return err
	})
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestWorkflowGrpcRollback(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	req := &busi.ReqGrpc{Amount: 30, TransInResult: "FAILURE"}
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqGrpc
		dtmgimp.MustProtoUnmarshal(data, &req)
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransOutRevertBSaga(wf.Context, &req)
			return err
		})
		_, err := busi.BusiCli.TransOutBSaga(wf.Context, &req)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransInRevertBSaga(wf.Context, &req)
			return err
		})
		_, err = busi.BusiCli.TransInBSaga(wf.Context, &req)
		return err
	})
	before := getBeforeBalances("mysql")
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err, dtmcli.ErrFailure)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assertSameBalance(t, before, "mysql")
}

func TestWorkflowMixed(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	gid := dtmimp.GetFuncName()
	err := workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqGrpc
		dtmgimp.MustProtoUnmarshal(data, &req)

		_, err := wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransOutRevertBSaga(wf.Context, &req)
			return err
		}).Do(func(bb *dtmcli.BranchBarrier) ([]byte, error) {
			return nil, bb.CallWithDB(dbGet().ToSQLDB(), func(tx *sql.Tx) error {
				return busi.SagaAdjustBalance(tx, busi.TransOutUID, int(-req.Amount), "")
			})
		})
		if err != nil {
			return err
		}
		wf.Context = metadata.NewOutgoingContext(wf.Context, metadata.Pairs("k1", "v1"))

		req2 := &busi.ReqHTTP{Amount: int(req.Amount / 2)}
		_, err = wf.NewBranch().OnCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req2).Post(Busi + "/TccBTransInConfirm")
			return err
		}).OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req2).Post(Busi + "/TccBTransInCancel")
			return err
		}).NewRequest().SetBody(req2).Post(Busi + "/TccBTransInTry")
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			return nil, busi.SagaAdjustBalance(db, busi.TransInUID, int(req.Amount/2), dtmcli.ResultSuccess)
		})
		return err
	})
	assert.Nil(t, err)
	before := getBeforeBalances("mysql")
	req := &busi.ReqGrpc{Amount: 30}
	err = workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Nil(t, err)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestWorkflowGrpcError(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	req := &busi.ReqGrpc{Amount: 30}
	gid := dtmimp.GetFuncName()
	busi.MainSwitch.TransOutResult.SetOnce("ERROR")
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqGrpc
		dtmgimp.MustProtoUnmarshal(data, &req)
		_, err := busi.BusiCli.TransOut(wf.NewBranchCtx(), &req)
		if err != nil {
			return err
		}
		_, err = busi.BusiCli.TransIn(wf.NewBranchCtx(), &req)
		return err
	})
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err)
	cronTransOnceForwardCron(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

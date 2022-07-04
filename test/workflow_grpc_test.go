/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"database/sql"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/workflow"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowGrpcSimple(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := &busi.ReqGrpc{Amount: 30, TransInResult: "FAILURE"}
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.BusiReq
		dtmgimp.MustProtoUnmarshal(data, &req)
		_, err := busi.BusiCli.TransOutBSaga(wf.NewBranchCtx(), &req)
		if err != nil {
			return err
		}
		_, err = busi.BusiCli.TransInBSaga(wf.NewBranchCtx(), &req)
		return err
	})
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err, dtmcli.ErrFailure)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	waitTransProcessed(gid)
}

func TestWorkflowGrpcNormal(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	req := &busi.BusiReq{Amount: 30, TransInResult: "FAILURE"}
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.BusiReq
		dtmgimp.MustProtoUnmarshal(data, &req)
		wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransOutRevertBSaga(wf.Context, &req)
			return err
		})
		_, err := busi.BusiCli.TransOutBSaga(wf.Context, &req)
		if err != nil {
			return err
		}
		wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransInRevertBSaga(wf.Context, &req)
			return err
		})
		_, err = busi.BusiCli.TransInBSaga(wf.Context, &req)
		return err
	})
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err, dtmcli.ErrFailure)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	waitTransProcessed(gid)
}

func TestWorkflowMixed(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := &busi.BusiReq{Amount: 30}
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.BusiReq
		dtmgimp.MustProtoUnmarshal(data, &req)

		wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransOutRevertBSaga(wf.Context, &req)
			return err
		})
		_, err := busi.BusiCli.TransOutBSaga(wf.Context, &req)
		if err != nil {
			return err
		}

		_, err = wf.NewBranch().OnBranchCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := busi.BusiCli.TransInConfirm(wf.Context, &req)
			return err
		}).OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			req2 := &busi.ReqHTTP{Amount: 30}
			_, err := wf.NewRequest().SetBody(req2).Post(Busi + "/TransInRevert")
			return err
		}).Do(func(bb *dtmcli.BranchBarrier) ([]byte, error) {
			err := busi.SagaAdjustBalance(dbGet().ToSQLDB(), busi.TransInUID, int(req.Amount), "")
			return nil, err
		})
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			return nil, busi.SagaAdjustBalance(db, busi.TransInUID, 0, dtmcli.ResultSuccess)
		})
		return err
	})
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Nil(t, err)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	waitTransProcessed(gid)
}

func TestWorkflowGrpcError(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	req := &busi.BusiReq{Amount: 30}
	gid := dtmimp.GetFuncName()
	busi.MainSwitch.TransOutResult.SetOnce("ERROR")
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.BusiReq
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
	go waitTransProcessed(gid)
	cronTransOnceForwardCron(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

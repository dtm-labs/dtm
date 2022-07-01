/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/workflow"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowNormal(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransIn")
		if err != nil {
			return err
		}
		return nil
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowSimpleResume(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	ongoingStep = 0

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		if fetchOngoingStep(0) {
			return dtmcli.ErrOngoing
		}
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		return err
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err)
	go waitTransProcessed(gid)
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowRollback(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)

	req := &busi.ReqHTTP{Amount: 30, TransInResult: dtmimp.ResultFailure}
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/SagaBTransOutCom")
			return err
		}).Do(func(bb *dtmcli.BranchBarrier) ([]byte, error) {
			return nil, bb.CallWithDB(dbGet().ToSQLDB(), func(tx *sql.Tx) error {
				return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "")
			})
		})
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			return bb.CallWithDB(dbGet().ToSQLDB(), func(tx *sql.Tx) error {
				return busi.SagaAdjustBalance(tx, busi.TransInUID, -req.Amount, "")
			})
		}).NewRequest().SetBody(req).Post(Busi + "/SagaBTransIn")
		if err != nil {
			return err
		}
		return nil
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
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

var ongoingStep = 0

func fetchOngoingStep(dest int) bool {
	c := ongoingStep
	logger.Debugf("ongoing step is: %d", c)
	if c == dest {
		ongoingStep++
		return true
	}
	return false
}

func TestWorkflowGrpcRollbackResume(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	gid := dtmimp.GetFuncName()
	ongoingStep = 0
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.BusiReq
		dtmgimp.MustProtoUnmarshal(data, &req)
		if fetchOngoingStep(0) {
			return dtmcli.ErrOngoing
		}
		wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			if fetchOngoingStep(4) {
				return dtmcli.ErrOngoing
			}
			_, err := busi.BusiCli.TransOutRevertBSaga(wf.Context, &req)
			return err
		})
		_, err := busi.BusiCli.TransOutBSaga(wf.Context, &req)
		if fetchOngoingStep(1) {
			return dtmcli.ErrOngoing
		}
		if err != nil {
			return err
		}
		wf.NewBranch().OnBranchRollback(func(bb *dtmcli.BranchBarrier) error {
			if fetchOngoingStep(3) {
				return dtmcli.ErrOngoing
			}
			_, err := busi.BusiCli.TransInRevertBSaga(wf.Context, &req)
			return err
		})
		_, err = busi.BusiCli.TransInBSaga(wf.Context, &req)
		if fetchOngoingStep(2) {
			return dtmcli.ErrOngoing
		}
		return err
	})
	req := &busi.BusiReq{Amount: 30, TransInResult: "FAILURE"}
	err := workflow.Execute(gid, gid, dtmgimp.MustProtoMarshal(req))
	assert.Error(t, err, dtmcli.ErrOngoing)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	// next cron will make a workflow submit, and do an additional write to chan, so make an additional read chan
	go waitTransProcessed(gid)
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestWorkflowXaAction(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		_, err := wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			return nil, busi.SagaAdjustBalance(db, busi.TransOutUID, -30, dtmcli.ResultSuccess)
		})
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			return nil, busi.SagaAdjustBalance(db, busi.TransInUID, 30, dtmcli.ResultSuccess)
		})
		return err
	})
	err := workflow.Execute(gid, gid, nil)
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowXaRollback(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		_, err := wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			return nil, busi.SagaAdjustBalance(db, busi.TransOutUID, -30, dtmcli.ResultSuccess)
		})
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			e := busi.SagaAdjustBalance(db, busi.TransInUID, 30, dtmcli.ResultSuccess)
			logger.FatalIfError(e)
			return nil, dtmcli.ErrFailure
		})
		return err
	})
	err := workflow.Execute(gid, gid, nil)
	assert.Equal(t, dtmcli.ErrFailure, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestWorkflowXaResume(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolGRPC)
	ongoingStep = 0
	gid := dtmimp.GetFuncName()
	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		_, err := wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			if fetchOngoingStep(0) {
				return nil, dtmcli.ErrOngoing
			}
			return nil, busi.SagaAdjustBalance(db, busi.TransOutUID, -30, dtmcli.ResultSuccess)
		})
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().DoXa(busi.BusiConf, func(db *sql.DB) ([]byte, error) {
			if fetchOngoingStep(1) {
				return nil, dtmcli.ErrOngoing
			}
			return nil, busi.SagaAdjustBalance(db, busi.TransInUID, 30, dtmcli.ResultSuccess)
		})
		if err != nil {
			return err
		}
		if fetchOngoingStep(2) {
			return dtmcli.ErrOngoing
		}

		return err
	})
	err := workflow.Execute(gid, gid, nil)
	assert.Equal(t, dtmcli.ErrOngoing, err)

	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusPrepared, getTransStatus(gid))
	// next cron will make a workflow submit, and do an additional write to chan, so make an additional read chan
	go waitTransProcessed(gid)
	cronTransOnceForwardNow(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowBranchConflict(t *testing.T) {
	gid := dtmimp.GetFuncName()
	store := dtmsvr.GetStore()
	now := time.Now()
	g := &storage.TransGlobalStore{
		Gid:          gid,
		Status:       dtmcli.StatusPrepared,
		NextCronTime: &now,
	}
	err := store.MaySaveNewTrans(g, []storage.TransBranchStore{
		{
			BranchID: "00",
			Op:       dtmimp.OpAction,
		},
	})
	assert.Nil(t, err)
	err = dtmimp.CatchP(func() {
		store.LockGlobalSaveBranches(gid, dtmcli.StatusPrepared, []storage.TransBranchStore{
			{BranchID: "00", Op: dtmimp.OpAction},
		}, -1)
	})
	assert.Error(t, err)
	store.ChangeGlobalStatus(g, StatusSucceed, []string{}, true)
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

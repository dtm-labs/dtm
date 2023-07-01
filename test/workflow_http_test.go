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
	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/dtm-labs/logger"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowNormal(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		wf.NewBranch().OnFinish(func(bb *dtmcli.BranchBarrier, isRollback bool) error {
			logger.Debugf("OnFinish isRollback: %v", isRollback)
			return nil
		})
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
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowRollback(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)

	req := &busi.ReqHTTP{Amount: 30, TransInResult: dtmimp.ResultFailure}
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		wf.NewBranch().OnFinish(func(bb *dtmcli.BranchBarrier, isRollback bool) error {
			logger.Debugf("OnFinish isRollback: %v", isRollback)
			return nil
		})
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
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
		_, err = wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return bb.CallWithDB(dbGet().ToSQLDB(), func(tx *sql.Tx) error {
				return busi.SagaAdjustBalance(tx, busi.TransInUID, -req.Amount, "")
			})
		}).NewRequest().SetBody(req).Post(Busi + "/SagaBTransIn")
		if err != nil {
			return err
		}
		return nil
	})
	before := getBeforeBalances("mysql")

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err, dtmcli.ErrFailure)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assertSameBalance(t, before, "mysql")
}

func TestWorkflowTcc(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransOutCancel")
			return err
		}).OnCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransOutConfirm")
			return err
		}).NewRequest().SetBody(req).Post(Busi + "/TccBTransOutTry")
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransInCancel")
			return err
		}).OnCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransInConfirm")
			return err
		}).NewRequest().SetBody(req).Post(Busi + "/TccBTransInTry")
		if err != nil {
			return err
		}
		return nil
	})

	before := getBeforeBalances("mysql")
	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Nil(t, err)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestWorkflowTccRollback(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, true)
	gid := dtmimp.GetFuncName()

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransOutCancel")
			return err
		}).OnCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransOutConfirm")
			return err
		}).NewRequest().SetBody(req).Post(Busi + "/TccBTransOutTry")
		if err != nil {
			return err
		}
		_, err = wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransInCancel")
			return err
		}).OnCommit(func(bb *dtmcli.BranchBarrier) error {
			_, err := wf.NewRequest().SetBody(req).Post(Busi + "/TccBTransInConfirm")
			return err
		}).NewRequest().SetBody(req).Post(Busi + "/TccBTransInTry")
		if err != nil {
			return err
		}
		return nil
	})

	before := getBeforeBalances("mysql")
	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assertSameBalance(t, before, "mysql")
}

func TestWorkflowError(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()
	busi.MainSwitch.TransOutResult.SetOnce("ERROR")

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		return err
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err)
	cronTransOnceForwardCron(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestWorkflowOngoing(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		return err
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err)
	cronTransOnceForwardCron(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

var resumeCounter int

func TestWorkflowResumeSkip(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()

	resumeCounter = 0
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")

	workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		wf.NewBranch().Do(func(bb *dtmcli.BranchBarrier) ([]byte, error) {
			logger.Infof("increase resume counter")
			resumeCounter++
			return nil, nil
		})
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		return err
	})

	err := workflow.Execute(gid, gid, dtmimp.MustMarshal(req))
	assert.Error(t, err)
	cronTransOnceForwardCron(t, gid, 1000)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, 1, resumeCounter)
}

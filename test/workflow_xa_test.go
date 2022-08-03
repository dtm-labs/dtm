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
	before := getBeforeBalances("mysql")
	err := workflow.Execute(gid, gid, nil)
	assert.Nil(t, err)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assertNotSameBalance(t, before, "mysql")
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
	before := getBeforeBalances("mysql")
	err := workflow.Execute(gid, gid, nil)
	assert.Equal(t, dtmcli.ErrFailure, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assertSameBalance(t, before, "mysql")
}

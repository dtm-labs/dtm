/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"context"
	"testing"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

var DtmServer = dtmutil.DefaultHTTPServer
var DtmGrpcServer = dtmutil.DefaultGrpcServer
var Busi = busi.Busi

func getTransStatus(gid string) string {
	return dtmsvr.GetTransGlobal(gid).Status
}

func getTrans(gid string) *dtmsvr.TransGlobal {
	return dtmsvr.GetTransGlobal(gid)
}

func getBranchesStatus(gid string) []string {
	branches := dtmsvr.GetStore().FindBranches(gid)
	status := []string{}
	for _, branch := range branches {
		status = append(status, branch.Status)
	}
	return status
}

func isSQLStore() bool {
	return conf.Store.Driver == config.Mysql || conf.Store.Driver == config.Postgres
}
func TestUpdateBranchAsync(t *testing.T) {
	if !isSQLStore() {
		return
	}
	conf.UpdateBranchSync = 0
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)

	gid := dtmimp.GetFuncName() + "-wf"
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	err = workflow.Register(gid, func(wf *workflow.Workflow, data []byte) error {
		_, err := busi.BusiCli.TransOut(wf.NewBranchCtx(), &busi.ReqGrpc{})
		// add additional data directly
		dtmimp.TransRegisterBranch(wf.TransBase, map[string]string{
			"branch_id": "01",
			"op":        "action",
			"status":    "succeed",
		}, "registerBranch")
		return err
	})
	assert.Nil(t, err)
	err = workflow.Execute(gid, gid, nil)
	assert.Nil(t, err)

	time.Sleep(dtmsvr.UpdateBranchAsyncInterval)

	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))

	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(gid))
	assert.Equal(t, StatusSucceed, getTransStatus(gid))

	conf.UpdateBranchSync = 1
}

func TestGrpcPanic(t *testing.T) {
	gid := dtmimp.GetFuncName()
	req := dtmgpb.DtmRequest{
		Gid: gid,
	}
	err := dtmgimp.MustGetGrpcConn(DtmGrpcServer, false).Invoke(context.Background(), "/dtmgimp.Dtm/"+"Submit", &req, nil)
	assert.Error(t, err)
}

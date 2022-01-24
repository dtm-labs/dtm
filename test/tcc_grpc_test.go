/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"context"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestTccGrpcNormal(t *testing.T) {
	req := busi.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutConfirm", busi.BusiGrpc+"/busi.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		return tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransIn", busi.BusiGrpc+"/busi.Busi/TransInConfirm", busi.BusiGrpc+"/busi.Busi/TransInRevert", r)
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))

}

func TestTccGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, true)
	err := dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOutTcc", busi.BusiGrpc+"/busi.Busi/TransOutConfirm", busi.BusiGrpc+"/busi.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		busi.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
		return tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransInTcc", busi.BusiGrpc+"/busi.Busi/TransInConfirm", busi.BusiGrpc+"/busi.Busi/TransInRevert", r)
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusAborting, getTransStatus(gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccGrpcNested(t *testing.T) {
	req := busi.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOutTcc", busi.BusiGrpc+"/busi.Busi/TransOutConfirm", busi.BusiGrpc+"/busi.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		return tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransInTccNested", busi.BusiGrpc+"/busi.Busi/TransInConfirm", busi.BusiGrpc+"/busi.Busi/TransInRevert", r)
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccGrpcType(t *testing.T) {
	_, err := dtmgrpc.TccFromGrpc(context.Background())
	assert.Error(t, err)
	logger.Debugf("expecting dtmutil.DefaultGrpcServer error")
	err = dtmgrpc.TccGlobalTransaction("-", "", func(tcc *dtmgrpc.TccGrpc) error { return nil })
	assert.Error(t, err)
}

func TestTccGrpcHeaders(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction2(dtmutil.DefaultGrpcServer, gid, func(tg *dtmgrpc.TccGrpc) {
		tg.BranchHeaders = map[string]string{
			"test_header": "test",
		}
		tg.WaitResult = true
	}, func(tcc *dtmgrpc.TccGrpc) error {
		data := &busi.BusiReq{Amount: 30}
		r := &emptypb.Empty{}
		return tcc.CallBranch(data, busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", "", r)
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(gid))

}

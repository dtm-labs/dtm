/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestSagaGrpcNormal(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSagaGrpc(gid, false, true)
	busi.MainSwitch.FailureReason.SetOnce("Insufficient balance")
	busi.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
	assert.Contains(t, getTrans(saga.Gid).RollbackReason, "Insufficient balance")
}

func TestSagaGrpcCurrent(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false).
		EnableConcurrent()
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcCurrentOrder(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false).
		EnableConcurrent().
		AddBranchOrder(1, []int{0})
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcCommittedOngoing(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSagaGrpc(gid, false, false)
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaGrpcNormalWait(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSagaGrpc(gid, false, false)
	saga.WaitResult = true
	saga.Submit()
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
}

func TestSagaGrpcEmptyUrl(t *testing.T) {
	saga := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, dtmimp.GetFuncName())
	req := busi.GenReqGrpc(30, false, false)
	saga.Add(busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutRevert", req)
	saga.Add("", busi.BusiGrpc+"/busi.Busi/TransInRevert", req)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

//nolint: unparam
func genSagaGrpc(gid string, outFailed bool, inFailed bool) *dtmgrpc.SagaGrpc {
	saga := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gid)
	req := busi.GenReqGrpc(30, outFailed, inFailed)
	saga.Add(busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutRevert", req)
	saga.Add(busi.BusiGrpc+"/busi.Busi/TransIn", busi.BusiGrpc+"/busi.Busi/TransInRevert", req)
	return saga
}

func TestSagaGrpcPassthroughHeadersYes(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gidYes)
	sagaYes.WaitResult = true
	sagaYes.PassthroughHeaders = []string{"test_header"}
	sagaYes.Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", nil)
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
}

func TestSagaGrpcWithGlobalTransRequestTimeout(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gid)
	saga.WaitResult = true
	saga.Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderNo", "", nil)
	saga.WithGlobalTransRequestTimeout(6)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gid)
}

func TestSagaGrpcOptionsRollbackWait(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSagaGrpc(gid, false, true)
	busi.MainSwitch.FailureReason.SetOnce("Insufficient balance")
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Insufficient balance")
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
	assert.Contains(t, getTrans(saga.Gid).RollbackReason, "Insufficient balance")
}

func TestSagaGrpcCronPassthroughHeadersYes(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gidYes)
	sagaYes.PassthroughHeaders = []string{"test_header"}
	sagaYes.Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", nil)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
	assert.Equal(t, StatusSubmitted, getTransStatus(gidYes))
	cronTransOnce(t, gidYes)
	assert.Equal(t, StatusSucceed, getTransStatus(gidYes))
}

func TestSagaGrpcPassthroughHeadersNo(t *testing.T) {
	gidNo := dtmimp.GetFuncName()
	sagaNo := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gidNo)
	sagaNo.WaitResult = true
	sagaNo.Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderNo", "", nil)
	err := sagaNo.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidNo)
}

func TestSagaGrpcHeaders(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gidYes).
		Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", nil)
	sagaYes.BranchHeaders = map[string]string{
		"test_header": "test",
	}
	sagaYes.WaitResult = true
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
}

func TestSagaGrpcCronHeaders(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmgrpc.NewSagaGrpc(dtmutil.DefaultGrpcServer, gidYes)
	sagaYes.BranchHeaders = map[string]string{
		"test_header": "test",
	}
	sagaYes.Add(busi.BusiGrpc+"/busi.Busi/TransOutHeaderYes", "", nil)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
	assert.Equal(t, StatusSubmitted, getTransStatus(gidYes))
	cronTransOnce(t, gidYes)
	assert.Equal(t, StatusSucceed, getTransStatus(gidYes))
}

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
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestSagaNormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaRollback(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
}

func TestSagaOngoingSucceed(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaFailed(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	busi.MainSwitch.TransOutRevertResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func TestSagaAbnormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	err = saga.Submit() // submit twice, ignored
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	err = saga.Submit()
	assert.Error(t, err) // a succeed trans can't accept submit
}

func TestSagaEmptyUrl(t *testing.T) {
	saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, dtmimp.GetFuncName())
	req := busi.GenTransReq(30, false, false)
	saga.Add(busi.Busi+"/TransOut", "", &req)
	saga.Add("", "", &req)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gid)
	req := busi.GenTransReq(30, outFailed, inFailed)
	saga.Add(busi.Busi+"/TransOut", busi.Busi+"/TransOutRevert", &req)
	saga.Add(busi.Busi+"/TransIn", busi.Busi+"/TransInRevert", &req)
	return saga
}

func genSaga1(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gid)
	req := busi.GenTransReq(30, outFailed, inFailed)
	saga.Add(busi.Busi+"/TransOut", busi.Busi+"/TransOutRevert", &req)
	return saga
}

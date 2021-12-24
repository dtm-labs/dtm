/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestSagaNormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOngoingSucceed(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaFailed(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func TestSagaAbnormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	err := saga.Submit()
	assert.Nil(t, err)
	err = saga.Submit() // submit twice, ignored
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	err = saga.Submit()
	assert.Error(t, err) // a succeed trans can't accept submit
}

func TestSagaEmptyUrl(t *testing.T) {
	saga := dtmcli.NewSaga(examples.DtmHttpServer, dtmimp.GetFuncName())
	req := examples.GenTransReq(30, false, false)
	saga.Add(examples.Busi+"/TransOut", "", &req)
	saga.Add("", "", &req)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(examples.DtmHttpServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	return saga
}

func genSaga1(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(examples.DtmHttpServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	return saga
}

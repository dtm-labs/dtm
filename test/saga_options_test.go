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

func TestSagaOptionsRetryOngoing(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is larger than RetryInterval
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	cronTransOnce(t, gid)
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsRetryError(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is less than 2*RetryInterval
	busi.MainSwitch.TransOutResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce(t, "")
	cronTransOnceForwardCron(t, gid, 360)
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsTimeout(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.TimeoutToFail = 1800
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnceForwardNow(t, gid, 3600)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
}

func TestSagaOptionsNormalWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Nil(t, err)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
}

func TestSagaOptionsCommittedOngoingWait(t *testing.T) {
	gid := dtmimp.GetFuncName()
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
	cronTransOnce(t, gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOptionsRollbackWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func TestSagaPassthroughHeadersYes(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gidYes)
	sagaYes.WaitResult = true
	sagaYes.PassthroughHeaders = []string{"test_header"}
	sagaYes.Add(busi.Busi+"/TransOutHeaderYes", "", nil)
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
}

func TestSagaCronPassthroughHeadersYes(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gidYes)
	sagaYes.PassthroughHeaders = []string{"test_header"}
	sagaYes.Add(busi.Busi+"/TransOutHeaderYes", "", nil)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
	assert.Equal(t, StatusSubmitted, getTransStatus(gidYes))
	cronTransOnce(t, gidYes)
	assert.Equal(t, StatusSucceed, getTransStatus(gidYes))
}

func TestSagaPassthroughHeadersNo(t *testing.T) {
	gidNo := dtmimp.GetFuncName()
	sagaNo := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gidNo)
	sagaNo.WaitResult = true
	sagaNo.Add(busi.Busi+"/TransOutHeaderNo", "", nil)
	err := sagaNo.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidNo)
}

func TestSagaHeaders(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gidYes)
	sagaYes.BranchHeaders = map[string]string{
		"test_header": "test",
	}
	sagaYes.WaitResult = true
	sagaYes.Add(busi.Busi+"/TransOutHeaderYes", "", nil)
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
}

func TestSagaHeadersYes1(t *testing.T) {
	gidYes := dtmimp.GetFuncName()
	sagaYes := dtmcli.NewSaga(dtmutil.DefaultHTTPServer, gidYes)
	sagaYes.BranchHeaders = map[string]string{
		"test_header": "test",
	}
	sagaYes.Add(busi.Busi+"/TransOutHeaderYes", "", nil)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	err := sagaYes.Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidYes)
	assert.Equal(t, StatusSubmitted, getTransStatus(gidYes))
	cronTransOnce(t, gidYes)
	assert.Equal(t, StatusSucceed, getTransStatus(gidYes))
}

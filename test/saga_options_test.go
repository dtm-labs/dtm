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
	"github.com/dtm-labs/dtm/examples"
	"github.com/stretchr/testify/assert"
)

func TestSagaOptionsRetryOngoing(t *testing.T) {
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is larger than RetryInterval
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsRetryError(t *testing.T) {
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is less than 2*RetryInterval
	examples.MainSwitch.TransOutResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnceForwardCron(360)
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsTimeout(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.TimeoutToFail = 1800
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnceForwardNow(3600)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
}

func TestSagaOptionsNormalWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Nil(t, err)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
}

func TestSagaOptionsCommittedOngoingWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOptionsRollbackWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

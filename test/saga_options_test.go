package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestSagaOptionsRetryOngoing(t *testing.T) {
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is larger than RetryInterval
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsRetryError(t *testing.T) {
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.RetryInterval = 150 // CronForwardDuration is less than 2*RetryInterval
	examples.MainSwitch.TransOutResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnceForwardCron(360)
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaOptionsTimeout(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.TimeoutToFail = 1800
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnceForwardNow(3600)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
}

func TestSagaOptionsNormalWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.SetOptions(&dtmimp.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOptionsCommittedOngoingWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.SetOptions(&dtmimp.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOptionsRollbackWait(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	saga.SetOptions(&dtmimp.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

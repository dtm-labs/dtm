package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestWaitSaga(t *testing.T) {

	sagaNormalWait(t)
	sagaCommittedPendingWait(t)
	sagaRollbackWait(t)
}

func sagaNormalWait(t *testing.T) {
	saga := genSaga("gid-noramlSagaWait", false, false)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaCommittedPendingWait(t *testing.T) {
	saga := genSaga("gid-committedPendingWait", false, false)
	examples.MainSwitch.TransOutResult.SetOnce("PENDING")
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func sagaRollbackWait(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2Wait", false, true)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

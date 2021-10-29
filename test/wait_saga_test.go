package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestWaitSaga(t *testing.T) {

	sagaNormalWait(t)
	sagaCommittedOngoingWait(t)
	sagaRollbackWait(t)
}

func sagaNormalWait(t *testing.T) {
	saga := genSaga("gid-noramlSagaWait", false, false)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaCommittedOngoingWait(t *testing.T) {
	saga := genSaga("gid-committedOngoingWait", false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func sagaRollbackWait(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2Wait", false, true)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Error(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

package dtmsvr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/examples"
)

func TestSagaWait(t *testing.T) {

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
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaCommittedPendingWait(t *testing.T) {
	saga := genSaga("gid-committedPendingWait", false, false)
	examples.MainSwitch.TransOutResult.SetOnce("PENDING")
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "prepared", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func sagaRollbackWait(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2Wait", false, true)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Error(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

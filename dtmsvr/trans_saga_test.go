package dtmsvr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestSaga(t *testing.T) {

	sagaNormal(t)
	sagaCommittedPending(t)
	sagaRollback(t)
}

func sagaNormal(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	examples.MainSwitch.TransOutResult.SetOnce("PENDING")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "prepared", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "aborting", getTransStatus(saga.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	dtmcli.Logf("beginning a saga test ---------------- %s", gid)
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	return saga
}

package test

import (
	"testing"

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
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
	err := saga.Submit() // 第二次提交
	assert.Error(t, err)
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	examples.MainSwitch.TransOutResult.SetOnce("PENDING")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	CronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
	err := saga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "aborting", getTransStatus(saga.Gid))
	CronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
	err = saga.Submit()
	assert.Error(t, err)
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	dtmcli.Logf("beginning a saga test ---------------- %s", gid)
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	return saga
}

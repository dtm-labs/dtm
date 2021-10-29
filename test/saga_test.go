package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestSaga(t *testing.T) {
	sagaNormal(t)
	sagaCommittedOngoing(t)
	sagaRollback(t)
	sagaRollback2(t)
	sagaTimeout(t)
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

func sagaCommittedOngoing(t *testing.T) {
	saga := genSaga("gid-committedOngoing", false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollback-saga", false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
	err = saga.Submit()
	assert.Error(t, err)
}

func sagaRollback2(t *testing.T) {
	saga := genSaga("gid-rollback-saga2", false, false)
	saga.TimeoutToFail = 1800
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	err := saga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(saga.Gid)
	cronTransOnceForwardNow(3600)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
}

func sagaTimeout(t *testing.T) {
	saga := genSaga("gid-timeout-saga", false, false)
	saga.TimeoutToFail = 1800
	examples.MainSwitch.TransOutResult.SetOnce("UNKOWN")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnceForwardNow(3600)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	dtmcli.Logf("beginning a saga test ---------------- %s", gid)
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	return saga
}

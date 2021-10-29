package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestCSaga(t *testing.T) {
	csagaNormal(t)
	csagaRollback(t)
	csagaRollback2(t)
	csagaCommittedOngoing(t)
}

func genCSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	dtmcli.Logf("beginning a concurrent saga test ---------------- %s", gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	csaga := dtmcli.NewSaga(examples.DtmServer, gid).
		Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req).
		Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req).
		EnableConcurrent()
	return csaga
}

func csagaNormal(t *testing.T) {
	csaga := genCSaga("gid-noraml-csaga", false, false)
	csaga.Submit()
	WaitTransProcessed(csaga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(csaga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(csaga.Gid))
}

func csagaRollback(t *testing.T) {
	csaga := genCSaga("gid-rollback-csaga", true, false)
	examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	err := csaga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(csaga.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(csaga.Gid))
	CronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(csaga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusFailed, dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(csaga.Gid))
	err = csaga.Submit()
	assert.Error(t, err)
}

func csagaRollback2(t *testing.T) {
	csaga := genCSaga("gid-rollback-csaga2", true, false)
	csaga.AddStepOrder(1, []int{0})
	err := csaga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(csaga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(csaga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusFailed, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(csaga.Gid))
	err = csaga.Submit()
	assert.Error(t, err)
}

func csagaCommittedOngoing(t *testing.T) {
	csaga := genCSaga("gid-committed-ongoing-csaga", false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	csaga.Submit()
	WaitTransProcessed(csaga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusDoing, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(csaga.Gid))
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(csaga.Gid))

	CronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(csaga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(csaga.Gid))
}

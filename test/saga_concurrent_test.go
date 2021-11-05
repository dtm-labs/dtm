package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func genSagaCon(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	dtmimp.Logf("beginning a concurrent saga test ---------------- %s", gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	sagaCon := dtmcli.NewSaga(examples.DtmServer, gid).
		Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req).
		Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req).
		EnableConcurrent()
	return sagaCon
}

func TestSagaConNormal(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), false, false)
	sagaCon.Submit()
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(sagaCon.Gid))
}

func TestSagaConRollback(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), true, false)
	examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	err := sagaCon.Submit()
	assert.Nil(t, err)
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(sagaCon.Gid))
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(sagaCon.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusFailed, dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	err = sagaCon.Submit()
	assert.Error(t, err)
}

func TestSagaConRollback2(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), true, false)
	sagaCon.AddBranchOrder(1, []int{0})
	err := sagaCon.Submit()
	assert.Nil(t, err)
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(sagaCon.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusFailed, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(sagaCon.Gid))
	err = sagaCon.Submit()
	assert.Error(t, err)
}

func TestSagaConCommittedOngoing(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	sagaCon.Submit()
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(sagaCon.Gid))

	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(sagaCon.Gid))
}

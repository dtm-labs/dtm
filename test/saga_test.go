package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestSagaNormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaOngoingSucceed(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaFailed(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce("ERROR")
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

func TestSagaAbnormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	err := saga.Submit()
	assert.Nil(t, err)
	err = saga.Submit() // submit twice, ignored
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	err = saga.Submit()
	assert.Error(t, err) // a succeed trans can't accept submit
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	return saga
}

func genSaga1(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	return saga
}

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcSaga(t *testing.T) {
	sagaGrpcNormal(t)
	sagaGrpcCommittedOngoing(t)
	sagaGrpcRollback(t)
}

func sagaGrpcNormal(t *testing.T) {
	saga := genSagaGrpc("gid-sagaGrpcNormal", false, false)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaGrpcCommittedOngoing(t *testing.T) {
	saga := genSagaGrpc("gid-committedOngoingGrpc", false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	CronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func sagaGrpcRollback(t *testing.T) {
	saga := genSagaGrpc("gid-rollbackSaga2Grpc", false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(saga.Gid))
	CronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

func genSagaGrpc(gid string, outFailed bool, inFailed bool) *dtmgrpc.SagaGrpc {
	dtmcli.Logf("beginning a grpc saga test ---------------- %s", gid)
	saga := dtmgrpc.NewSaga(examples.DtmGrpcServer, gid)
	req := dtmcli.MustMarshal(examples.GenTransReq(30, outFailed, inFailed))
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutRevert", req)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInRevert", req)
	return saga
}

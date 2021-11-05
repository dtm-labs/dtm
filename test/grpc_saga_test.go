package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcSagaNormal(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestGrpcSagaCommittedOngoing(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
}

func TestGrpcSagaRollback(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

func genSagaGrpc(gid string, outFailed bool, inFailed bool) *dtmgrpc.SagaGrpc {
	dtmimp.Logf("beginning a grpc saga test ---------------- %s", gid)
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, gid)
	req := examples.GenBusiReq(30, outFailed, inFailed)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutRevert", req)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInRevert", req)
	return saga
}

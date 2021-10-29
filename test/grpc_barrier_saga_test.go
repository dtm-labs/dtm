package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcBarrierSaga(t *testing.T) {

	grpcSagaBarrierNormal(t)
	grpcSagaBarrierRollback(t)
}

func grpcSagaBarrierNormal(t *testing.T) {
	req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30})
	saga := dtmgrpc.NewSaga(examples.DtmGrpcServer, "grpcSagaBarrierNormal").
		Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func grpcSagaBarrierRollback(t *testing.T) {
	req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure})
	saga := dtmgrpc.NewSaga(examples.DtmGrpcServer, "grpcSagaBarrierRollback").
		Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

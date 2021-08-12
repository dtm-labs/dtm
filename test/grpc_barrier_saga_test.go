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
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
}

func grpcSagaBarrierRollback(t *testing.T) {
	req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"})
	saga := dtmgrpc.NewSaga(examples.DtmGrpcServer, "grpcSagaBarrierRollback").
		Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	err := saga.Submit()
	e2p(err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

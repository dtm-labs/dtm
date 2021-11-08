package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestSagaGrpcBarrierNormal(t *testing.T) {
	saga := genSagaGrpcBarrier(dtmimp.GetFuncName(), false, false)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaGrpcBarrierRollback(t *testing.T) {
	saga := genSagaGrpcBarrier(dtmimp.GetFuncName(), false, true)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func genSagaGrpcBarrier(gid string, outFailed bool, inFailed bool) *dtmgrpc.SagaGrpc {
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, gid)
	req := examples.GenBusiReq(30, outFailed, inFailed)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	return saga
}

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcSagaBarrierNormal(t *testing.T) {
	req := examples.GenBusiReq(30, false, false)
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, dtmimp.GetFuncName()).
		Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestGrpcSagaBarrierRollback(t *testing.T) {
	req := examples.GenBusiReq(30, false, true)
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, dtmimp.GetFuncName()).
		Add(examples.BusiGrpc+"/examples.Busi/TransOutBSaga", examples.BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransInBSaga", examples.BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusSucceed, dtmcli.StatusFailed}, getBranchesStatus(saga.Gid))
}

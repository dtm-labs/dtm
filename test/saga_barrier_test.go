package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestSagaBarrierNormal(t *testing.T) {
	req := &examples.TransReq{Amount: 30}
	saga := dtmcli.NewSaga(DtmServer, dtmimp.GetFuncName()).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
	dtmimp.Logf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaBarrierRollback(t *testing.T) {
	saga := dtmcli.NewSaga(DtmServer, dtmimp.GetFuncName()).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", &examples.TransReq{Amount: 30}).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", &examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure})
	dtmimp.Logf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
}

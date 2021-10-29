package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestBarrierSaga(t *testing.T) {

	sagaBarrierNormal(t)
	sagaBarrierRollback(t)
}

func sagaBarrierNormal(t *testing.T) {
	req := &examples.TransReq{Amount: 30}
	saga := dtmcli.NewSaga(DtmServer, "sagaBarrierNormal").
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
	dtmcli.Logf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
}

func sagaBarrierRollback(t *testing.T) {
	saga := dtmcli.NewSaga(DtmServer, "sagaBarrierRollback").
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", &examples.TransReq{Amount: 30}).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", &examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure})
	dtmcli.Logf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(saga.Gid))
}

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcTcc(t *testing.T) {
	// tccGrpcType(t)
	// tccGrpcNormal(t)
	// tccGrpcNested(t)
	tccGrpcRollback(t)
}

func tccGrpcType(t *testing.T) {
	_, err := dtmgrpc.TccFromRequest(&dtmgrpc.BusiRequest{Info: &dtmgrpc.BranchInfo{}})
	assert.Error(t, err)
	dtmcli.Logf("expecting dtmgrpcserver error")
	err = dtmgrpc.TccGlobalTransaction("-", "", func(tcc *dtmgrpc.TccGrpc) error { return nil })
	assert.Error(t, err)
}
func tccGrpcNormal(t *testing.T) {
	data := dtmcli.MustMarshal(&examples.TransReq{Amount: 30})
	gid := "tccGrpcNormal"
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		_, err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert")
		assert.Nil(t, err)
		_, err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert")
		return err
	})
	assert.Nil(t, err)
}

func tccGrpcNested(t *testing.T) {
	data := dtmcli.MustMarshal(&examples.TransReq{Amount: 30})
	gid := "tccGrpcNested"
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		_, err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert")
		assert.Nil(t, err)
		_, err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransInTccNested", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert")
		return err
	})
	assert.Nil(t, err)
}

func tccGrpcRollback(t *testing.T) {
	gid := "tccGrpcRollback"
	data := dtmcli.MustMarshal(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"})
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		_, err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert")
		assert.Nil(t, err)
		examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
		_, err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransInTcc", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert")
		return err
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, "aborting", getTransStatus(gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "failed", getTransStatus(gid))
}

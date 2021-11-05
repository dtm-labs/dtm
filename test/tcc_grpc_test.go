package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestTccGrpcType(t *testing.T) {
	_, err := dtmgrpc.TccFromGrpc(context.Background())
	assert.Error(t, err)
	dtmimp.Logf("expecting dtmgrpcserver error")
	err = dtmgrpc.TccGlobalTransaction("-", "", func(tcc *dtmgrpc.TccGrpc) error { return nil })
	assert.Error(t, err)
}
func TestTccGrpcNormal(t *testing.T) {
	data := &examples.BusiReq{Amount: 30}
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
		return err
	})
	assert.Nil(t, err)
}

func TestGrpcTestNested(t *testing.T) {
	data := &examples.BusiReq{Amount: 30}
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransInTccNested", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
		return err
	})
	assert.Nil(t, err)
}

func TestTccGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	data := &examples.BusiReq{Amount: 30, TransInResult: dtmcli.ResultFailure}
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
		err = tcc.CallBranch(data, examples.BusiGrpc+"/examples.Busi/TransInTcc", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
		return err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusAborting, getTransStatus(gid))
	cronTransOnce()
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

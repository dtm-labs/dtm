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

func TestTccGrpcNormal(t *testing.T) {
	req := examples.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		return tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))

}

func TestTccGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	req := examples.GenBusiReq(30, false, true)
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
		return tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInTcc", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusAborting, getTransStatus(gid))
	cronTransOnce()
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccGrpcNested(t *testing.T) {
	req := examples.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutTcc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Nil(t, err)
		return tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInTccNested", examples.BusiGrpc+"/examples.Busi/TransInConfirm", examples.BusiGrpc+"/examples.Busi/TransInRevert", r)
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccGrpcType(t *testing.T) {
	_, err := dtmgrpc.TccFromGrpc(context.Background())
	assert.Error(t, err)
	dtmimp.Logf("expecting dtmgrpcserver error")
	err = dtmgrpc.TccGlobalTransaction("-", "", func(tcc *dtmgrpc.TccGrpc) error { return nil })
	assert.Error(t, err)
}

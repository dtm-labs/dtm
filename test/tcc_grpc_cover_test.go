package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestTccGrpcCoverNotConnected(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction("localhost:01", gid, func(tcc *dtmgrpc.TccGrpc) error {
		return nil
	})
	assert.Error(t, err)
}

func TestTccGrpcCoverPanic(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmimp.CatchP(func() {
		_ = dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			panic("user panic")
		})
		assert.FailNow(t, "not executed")
	})
	assert.Contains(t, err.Error(), "user panic")
}

func TestTccGrpcCoverCallBranch(t *testing.T) {
	req := examples.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(examples.DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {

		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, "not_exists://abc", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Error(t, err)

		tcc.Dtm = "localhost:01"
		err = tcc.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutConfirm", examples.BusiGrpc+"/examples.Busi/TransOutRevert", r)
		assert.Error(t, err)

		return err
	})
	assert.Error(t, err)
}

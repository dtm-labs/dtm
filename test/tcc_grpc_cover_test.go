package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
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
		_ = dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			panic("user panic")
		})
		assert.FailNow(t, "not executed")
	})
	assert.Contains(t, err.Error(), "user panic")
	waitTransProcessed(gid)
}

func TestTccGrpcCoverCallBranch(t *testing.T) {
	req := busi.GenBusiReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction(dtmutil.DefaultGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {

		r := &emptypb.Empty{}
		err := tcc.CallBranch(req, "not_exists://abc", busi.BusiGrpc+"/busi.Busi/TransOutConfirm", busi.BusiGrpc+"/busi.Busi/TransOutRevert", r)
		assert.Error(t, err)

		tcc.Dtm = "localhost:01"
		err = tcc.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutConfirm", busi.BusiGrpc+"/busi.Busi/TransOutRevert", r)
		assert.Error(t, err)

		return err
	})
	assert.Error(t, err)
	cronTransOnceForwardNow(t, gid, 300)
}

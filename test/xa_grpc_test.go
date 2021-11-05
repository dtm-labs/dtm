package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestXaGrpcType(t *testing.T) {
	_, err := dtmgrpc.XaGrpcFromRequest(context.Background())
	assert.Error(t, err)

	err = examples.XaGrpcClient.XaLocalTransaction(context.Background(), nil, nil)
	assert.Error(t, err)

	err = dtmimp.CatchP(func() {
		examples.XaGrpcClient.XaGlobalTransaction("id1", func(xa *dtmgrpc.XaGrpc) error { panic(fmt.Errorf("hello")) })
	})
	assert.Error(t, err)
}

func TestXaGrpcLocalError(t *testing.T) {
	xc := examples.XaGrpcClient
	err := xc.XaGlobalTransaction(dtmimp.GetFuncName(), func(xa *dtmgrpc.XaGrpc) error {
		return fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
}

func TestXaGrpcNormal(t *testing.T) {
	xc := examples.XaGrpcClient
	gid := dtmimp.GetFuncName()
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := &examples.BusiReq{Amount: 30}
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		err = xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa", r)
		return err
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestXaGrpcRollback(t *testing.T) {
	xc := examples.XaGrpcClient
	gid := dtmimp.GetFuncName()
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := &examples.BusiReq{Amount: 30, TransInResult: dtmcli.ResultFailure}
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		err = xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa", r)
		return err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

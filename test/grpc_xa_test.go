package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcXa(t *testing.T) {
	if config.DB["driver"] != "mysql" {
		return
	}
	xaGrpcType(t)
	xaGrpcLocalError(t)
	xaGrpcNormal(t)
	xaGrpcRollback(t)
}

func xaGrpcType(t *testing.T) {
	_, err := dtmgrpc.XaGrpcFromRequest(&dtmgrpc.BusiRequest{Info: &dtmgrpc.BranchInfo{}})
	assert.Error(t, err)

	err = examples.XaGrpcClient.XaLocalTransaction(&dtmgrpc.BusiRequest{Info: &dtmgrpc.BranchInfo{}}, nil)
	assert.Error(t, err)

	err = dtmcli.CatchP(func() {
		examples.XaGrpcClient.XaGlobalTransaction("id1", func(xa *dtmgrpc.XaGrpc) error { panic(fmt.Errorf("hello")) })
	})
	assert.Error(t, err)
}

func xaGrpcLocalError(t *testing.T) {
	xc := examples.XaGrpcClient
	err := xc.XaGlobalTransaction("xaGrpcLocalError", func(xa *dtmgrpc.XaGrpc) error {
		return fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
}

func xaGrpcNormal(t *testing.T) {
	xc := examples.XaGrpcClient
	gid := "xaGrpcNormal"
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := dtmcli.MustMarshal(examples.GenTransReq(30, false, false))
		_, err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa")
		if err != nil {
			return err
		}
		_, err = xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa")
		return err
	})
	assert.Equal(t, nil, err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(gid))
}

func xaGrpcRollback(t *testing.T) {
	xc := examples.XaGrpcClient
	gid := "xaGrpcRollback"
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure})
		_, err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa")
		if err != nil {
			return err
		}
		_, err = xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa")
		return err
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
}

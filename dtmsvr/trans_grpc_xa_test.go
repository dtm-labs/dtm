package dtmsvr

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
	// xaGrpcLocalError(t)
	xaGrpcNormal(t)
	xaGrpcRollback(t)
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
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(gid))
}

func xaGrpcRollback(t *testing.T) {
	xc := examples.XaGrpcClient
	gid := "xaGrpcRollback"
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"})
		_, err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa")
		if err != nil {
			return err
		}
		_, err = xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa")
		return err
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"succeed", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}

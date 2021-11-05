package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestMsgGrpcNormal(t *testing.T) {
	msg := genGrpcMsg(dtmimp.GetFuncName())
	err := msg.Submit()
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgGrpcOngoingSuccess(t *testing.T) {
	msg := genGrpcMsg(dtmimp.GetFuncName())
	err := msg.Prepare("")
	assert.Nil(t, err)
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnce()
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgGrpcOngoingFailed(t *testing.T) {
	msg := genGrpcMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(180)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

func genGrpcMsg(gid string) *dtmgrpc.MsgGrpc {
	req := &examples.BusiReq{Amount: 30}
	msg := dtmgrpc.NewMsgGrpc(examples.DtmGrpcServer, gid).
		Add(examples.BusiGrpc+"/examples.Busi/TransOut", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransIn", req)
	msg.QueryPrepared = fmt.Sprintf("%s/examples.Busi/CanSubmit", examples.BusiGrpc)
	return msg
}

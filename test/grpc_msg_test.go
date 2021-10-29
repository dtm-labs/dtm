package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcMsg(t *testing.T) {
	grpcMsgNormal(t)
	grpcMsgOngoing(t)
}

func grpcMsgNormal(t *testing.T) {
	msg := genGrpcMsg("grpc-msg-normal")
	err := msg.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func grpcMsgOngoing(t *testing.T) {
	msg := genGrpcMsg("grpc-msg-pending")
	err := msg.Prepare(fmt.Sprintf("%s/examples.Busi/CanSubmit", examples.BusiGrpc))
	assert.Nil(t, err)
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnce()
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func genGrpcMsg(gid string) *dtmgrpc.MsgGrpc {
	req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30})
	return dtmgrpc.NewMsgGrpc(examples.DtmGrpcServer, gid).
		Add(examples.BusiGrpc+"/examples.Busi/TransOut", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransIn", req)

}

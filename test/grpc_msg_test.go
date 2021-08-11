package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestGrpcMsg(t *testing.T) {
	grpcMsgNormal(t)
	grpcMsgPending(t)
}

func grpcMsgNormal(t *testing.T) {
	msg := genGrpcMsg("grpc-msg-normal")
	err := msg.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func grpcMsgPending(t *testing.T) {
	msg := genGrpcMsg("grpc-msg-pending")
	err := msg.Prepare(fmt.Sprintf("%s/examples.Busi/CanSubmit", examples.BusiGrpc))
	assert.Nil(t, err)
	examples.MainSwitch.CanSubmitResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "submitted", getTransStatus(msg.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func genGrpcMsg(gid string) *dtmgrpc.MsgGrpc {
	req := dtmcli.MustMarshal(&examples.TransReq{Amount: 30})
	return dtmgrpc.NewMsgGrpc(examples.DtmGrpcServer, gid).
		Add(examples.BusiGrpc+"/examples.Busi/TransOut", req).
		Add(examples.BusiGrpc+"/examples.Busi/TransIn", req)

}

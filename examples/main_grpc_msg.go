package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

// MsgGrpcSetup 1
func MsgGrpcSetup(app *gin.Engine) {

}

// MsgGrpcFireRequest 1
func MsgGrpcFireRequest() string {
	req := dtmcli.MustMarshal(&TransReq{Amount: 30})
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, dtmcli.MustGenGid(DtmServer)).
		Add(BusiPb+"/examples.Busi/TransOut", req).
		Add(BusiPb+"/examples.Busi/TransIn", req)
	err := msg.Submit()
	e2p(err)
	return msg.Gid
}

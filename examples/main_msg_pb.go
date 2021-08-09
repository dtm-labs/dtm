package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	dtmpb "github.com/yedf/dtm/dtmpb"
)

// MsgPbSetup 1
func MsgPbSetup(app *gin.Engine) {

}

// MsgPbFireRequest 1
func MsgPbFireRequest() string {
	req := dtmcli.MustMarshal(&TransReq{Amount: 30})
	msg := dtmpb.NewMsgGrpc(DtmGrpcServer, dtmcli.MustGenGid(DtmServer)).
		Add(BusiPb+"/examples.Busi/TransOut", req).
		Add(BusiPb+"/examples.Busi/TransIn", req)
	err := msg.Submit()
	e2p(err)
	return msg.Gid
}

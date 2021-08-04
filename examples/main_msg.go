package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// MsgSetup 1
func MsgSetup(app *gin.Engine) {
}

// MsgFireRequest 1
func MsgFireRequest() string {
	common.Logf("a busi transaction begin")
	req := &TransReq{Amount: 30}
	msg := dtmcli.NewMsg(DtmServer, dtmcli.MustGenGid(DtmServer)).
		Add(Busi+"/TransOut", req).
		Add(Busi+"/TransIn", req)
	err := msg.Prepare(Busi + "/TransQuery")
	e2p(err)
	common.Logf("busi trans submit")
	err = msg.Submit()
	e2p(err)
	return msg.Gid
}

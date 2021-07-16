package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtmcli"
)

// MsgSetup 1
func MsgSetup(app *gin.Engine) {
}

// MsgFireRequest 1
func MsgFireRequest() {
	logrus.Printf("a busi transaction begin")
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	msg := dtmcli.NewMsg(DtmServer).
		Add(Busi+"/TransOut", req).
		Add(Busi+"/TransIn", req)
	err := msg.Prepare(Busi + "/TransQuery")
	e2p(err)
	logrus.Printf("busi trans submit")
	err = msg.Submit()
	e2p(err)
}

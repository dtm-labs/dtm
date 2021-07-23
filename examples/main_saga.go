package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtmcli"
)

// SagaSetup 1
func SagaSetup(app *gin.Engine) {
}

// SagaFireRequest 1
func SagaFireRequest() string {
	logrus.Printf("a saga busi transaction begin")
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
		Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
		Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
	logrus.Printf("saga busi trans submit")
	err := saga.Submit()
	logrus.Printf("result gid is: %s", saga.Gid)
	e2p(err)
	return saga.Gid
}

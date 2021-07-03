package examples

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtmcli"
)

func SagaMain() {
	app := BaseAppNew()
	BaseAppSetup(app)
	TccSetup(app)
	go BaseAppStart(app)
	time.Sleep(100 * time.Millisecond)
	TccFireRequest()
	time.Sleep(1000 * time.Second)
}

func SagaSetup(app *gin.Engine) {
}

func SagaFireRequest() {
	logrus.Printf("a saga busi transaction begin")
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtmcli.SagaNew(DtmServer).
		Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
		Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
	logrus.Printf("saga busi trans submit")
	err := saga.Submit()
	logrus.Printf("result gid is: %s", saga.Gid)
	e2p(err)
}

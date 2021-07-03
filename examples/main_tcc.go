package examples

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm"
)

func TccMain() {
	app := BaseAppNew()
	BaseAppSetup(app)
	TccSetup(app)
	go BaseAppStart(app)
	time.Sleep(100 * time.Millisecond)
	TccFireRequest()
	time.Sleep(1000 * time.Second)
}

func TccSetup(app *gin.Engine) {
}

func TccFireRequest() {
	logrus.Printf("tcc transaction begin")
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	tcc := dtm.TccNew(DtmServer).
		Add(Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert", req).
		Add(Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransOutRevert", req)
	logrus.Printf("tcc trans commit")
	err := tcc.Commit()
	e2p(err)
}

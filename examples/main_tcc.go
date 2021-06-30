package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm"
	"github.com/yedf/dtm/common"
)

// 事务参与者的服务地址
const TccBusiApi = "/api/busi_tcc"

var TccBusi = fmt.Sprintf("http://localhost:%d%s", TccBusiPort, TccBusiApi)

func TccMain() {
	go TccStartSvr()
	TccFireRequest()
	time.Sleep(1000 * time.Second)
}

func TccStartSvr() {
	logrus.Printf("tcc examples starting")
	app := common.GetGinApp()
	TccAddRoute(app)
	app.Run(fmt.Sprintf(":%d", TccBusiPort))
}

func TccFireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	tcc := dtm.TccNew(DtmServer, gid).
		Add(TccBusi+"/TransOutTry", TccBusi+"/TransOutConfirm", TccBusi+"/TransOutCancel", req).
		Add(TccBusi+"/TransInTry", TccBusi+"/TransInConfirm", TccBusi+"/TransOutCancel", req)
	logrus.Printf("busi trans commit")
	err := tcc.Commit()
	e2p(err)
}

// api

func TccAddRoute(app *gin.Engine) {
	app.POST(TccBusiApi+"/TransInTry", common.WrapHandler(tccTransInTry))
	app.POST(TccBusiApi+"/TransInConfirm", common.WrapHandler(tccTransInConfirm))
	app.POST(TccBusiApi+"/TransInCancel", common.WrapHandler(tccTransCancel))
	app.POST(TccBusiApi+"/TransOutTry", common.WrapHandler(tccTransOutTry))
	app.POST(TccBusiApi+"/TransOutConfirm", common.WrapHandler(tccTransOutConfirm))
	app.POST(TccBusiApi+"/TransOutCancel", common.WrapHandler(tccTransOutCancel))
	logrus.Printf("examples listening at %d", TccBusiPort)
}

var TccTransInTryResult = ""
var TccTransOutTryResult = ""
var TccTransInCancelResult = ""
var TccTransOutCancelResult = ""
var TccTransInConfirmResult = ""
var TccTransOutConfirmResult = ""

func tccTransInTry(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransInTryResult, req.TransInResult, "SUCCESS")
	logrus.Printf("%s TransInTry: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func tccTransInConfirm(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransInConfirmResult, "SUCCESS")
	logrus.Printf("%s tccTransInConfirm: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func tccTransCancel(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransInCancelResult, "SUCCESS")
	logrus.Printf("%s tccTransCancel: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func tccTransOutTry(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransOutTryResult, req.TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func tccTransOutConfirm(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransOutConfirmResult, "SUCCESS")
	logrus.Printf("%s TransOutConfirm: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func tccTransOutCancel(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TccTransOutCancelResult, "SUCCESS")
	logrus.Printf("%s tccTransOutCancel: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

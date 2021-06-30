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
const MsgBusiApi = "/api/busi_msg"

var MsgBusi = fmt.Sprintf("http://localhost:%d%s", MsgBusiPort, MsgBusiApi)

func MsgMain() {
	go MsgStartSvr()
	MsgFireRequest()
	time.Sleep(1000 * time.Second)
}

func MsgStartSvr() {
	logrus.Printf("msg examples starting")
	app := common.GetGinApp()
	MsgAddRoute(app)
	app.Run(fmt.Sprintf(":%d", MsgBusiPort))
}

func MsgFireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	msg := dtm.MsgNew(DtmServer, gid).
		Add(MsgBusi+"/TransOut", req).
		Add(MsgBusi+"/TransIn", req)
	err := msg.Prepare(MsgBusi + "/TransQuery")
	e2p(err)
	logrus.Printf("busi trans commit")
	err = msg.Commit()
	e2p(err)
}

// api

func MsgAddRoute(app *gin.Engine) {
	app.POST(MsgBusiApi+"/TransIn", common.WrapHandler(msgTransIn))
	app.POST(MsgBusiApi+"/TransOut", common.WrapHandler(MsgTransOut))
	app.GET(MsgBusiApi+"/TransQuery", common.WrapHandler(msgTransQuery))
	logrus.Printf("examples msg listening at %d", MsgBusiPort)
}

var MsgTransInResult = ""
var MsgTransOutResult = ""
var MsgTransQueryResult = ""

func msgTransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(MsgTransInResult, req.TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func MsgTransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(MsgTransOutResult, req.TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func msgTransQuery(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	res := common.OrString(MsgTransQueryResult, "SUCCESS")
	return M{"result": res}, nil
}

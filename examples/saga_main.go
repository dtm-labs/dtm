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
const SagaBusiPort = 8081
const SagaBusiApi = "/api/busi_saga"

var SagaBusi = fmt.Sprintf("http://localhost:%d%s", SagaBusiPort, SagaBusiApi)

func SagaMain() {
	go SagaStartSvr()
	sagaFireRequest()
	time.Sleep(1000 * time.Second)
}

func SagaStartSvr() {
	logrus.Printf("saga examples starting")
	app := common.GetGinApp()
	AddRoute(app)
	app.Run(":8081")
}

func sagaFireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtm.SagaNew(DtmServer, gid, SagaBusi+"/TransQuery")

	saga.Add(SagaBusi+"/TransOut", SagaBusi+"/TransOutCompensate", req)
	saga.Add(SagaBusi+"/TransIn", SagaBusi+"/TransInCompensate", req)
	err := saga.Prepare()
	common.PanicIfError(err)
	logrus.Printf("busi trans commit")
	err = saga.Commit()
	common.PanicIfError(err)
}

// api

func AddRoute(app *gin.Engine) {
	app.POST(SagaBusiApi+"/TransIn", common.WrapHandler(TransIn))
	app.POST(SagaBusiApi+"/TransInCompensate", common.WrapHandler(TransInCompensate))
	app.POST(SagaBusiApi+"/TransOut", common.WrapHandler(TransOut))
	app.POST(SagaBusiApi+"/TransOutCompensate", common.WrapHandler(TransOutCompensate))
	app.GET(SagaBusiApi+"/TransQuery", common.WrapHandler(TransQuery))
	logrus.Printf("examples listening at %d", SagaBusiPort)
}

type M = map[string]interface{}

var TransInResult = ""
var TransOutResult = ""
var TransInCompensateResult = ""
var TransOutCompensateResult = ""
var TransQueryResult = ""

func transReqFromContext(c *gin.Context) *TransReq {
	req := TransReq{}
	err := c.BindJSON(&req)
	common.PanicIfError(err)
	return &req
}

func TransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TransInResult, req.TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransInCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TransInCompensateResult, "SUCCESS")
	logrus.Printf("%s TransInCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TransOutResult, req.TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransOutCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(TransOutCompensateResult, "SUCCESS")
	logrus.Printf("%s TransOutCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransQuery(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	res := common.OrString(TransQueryResult, "SUCCESS")
	return M{"result": res}, nil
}

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
	SagaFireRequest()
	time.Sleep(1000 * time.Second)
}

func SagaStartSvr() {
	logrus.Printf("saga examples starting")
	app := common.GetGinApp()
	SagaAddRoute(app)
	app.Run(fmt.Sprintf(":%d", SagaBusiPort))
}

func SagaFireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtm.SagaNew(DtmServer, gid).
		Add(SagaBusi+"/TransOut", SagaBusi+"/TransOutCompensate", req).
		Add(SagaBusi+"/TransIn", SagaBusi+"/TransInCompensate", req)
	err := saga.Prepare(SagaBusi + "/TransQuery")
	e2p(err)
	logrus.Printf("busi trans commit")
	err = saga.Commit()
	e2p(err)
}

// api

func SagaAddRoute(app *gin.Engine) {
	app.POST(SagaBusiApi+"/TransIn", common.WrapHandler(sagaTransIn))
	app.POST(SagaBusiApi+"/TransInCompensate", common.WrapHandler(sagaTransInCompensate))
	app.POST(SagaBusiApi+"/TransOut", common.WrapHandler(SagaTransOut))
	app.POST(SagaBusiApi+"/TransOutCompensate", common.WrapHandler(sagaTransOutCompensate))
	app.GET(SagaBusiApi+"/TransQuery", common.WrapHandler(sagaTransQuery))
	logrus.Printf("examples listening at %d", SagaBusiPort)
}

var SagaTransInResult = ""
var SagaTransOutResult = ""
var SagaTransInCompensateResult = ""
var SagaTransOutCompensateResult = ""
var SagaTransQueryResult = ""

func sagaTransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(SagaTransInResult, req.TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func sagaTransInCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(SagaTransInCompensateResult, "SUCCESS")
	logrus.Printf("%s TransInCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func SagaTransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(SagaTransOutResult, req.TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func sagaTransOutCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	res := common.OrString(SagaTransOutCompensateResult, "SUCCESS")
	logrus.Printf("%s TransOutCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func sagaTransQuery(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	res := common.OrString(SagaTransQueryResult, "SUCCESS")
	return M{"result": res}, nil
}

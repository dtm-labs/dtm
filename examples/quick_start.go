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
const startBusiPort = 8084
const startBusiApi = "/api/busi_start"

var startBusi = fmt.Sprintf("http://localhost:%d%s", startBusiPort, startBusiApi)

func startMain() {
	go startStartSvr()
	startFireRequest()
	time.Sleep(1000 * time.Second)
}

func startStartSvr() {
	logrus.Printf("saga examples starting")
	app := common.GetGinApp()
	startAddRoute(app)
	app.Run(fmt.Sprintf(":%d", SagaBusiPort))
}

func startFireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtm.SagaNew(DtmServer, gid).
		Add(startBusi+"/TransOut", startBusi+"/TransOutCompensate", req).
		Add(startBusi+"/TransIn", startBusi+"/TransInCompensate", req)
	logrus.Printf("busi trans commit")
	err := saga.Commit()
	e2p(err)
}

func startAddRoute(app *gin.Engine) {
	app.POST(SagaBusiApi+"/TransIn", common.WrapHandler(startTransIn))
	app.POST(SagaBusiApi+"/TransInCompensate", common.WrapHandler(startTransInCompensate))
	app.POST(SagaBusiApi+"/TransOut", common.WrapHandler(startTransOut))
	app.POST(SagaBusiApi+"/TransOutCompensate", common.WrapHandler(startTransOutCompensate))
	logrus.Printf("examples listening at %d", startBusiPort)
}

func startTransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	logrus.Printf("%s TransIn: %v result: %s", gid, req, req.TransInResult)
	return M{"result": req.TransInResult}, nil
}

func startTransInCompensate(c *gin.Context) (interface{}, error) {
	return M{"result": "SUCCESS"}, nil
}

func startTransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := transReqFromContext(c)
	logrus.Printf("%s TransOut: %v result: %s", gid, req, req.TransOutResult)
	return M{"result": req.TransOutResult}, nil
}

func startTransOutCompensate(c *gin.Context) (interface{}, error) {
	return M{"result": "SUCCESS"}, nil
}

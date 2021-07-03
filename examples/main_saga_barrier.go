package examples

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
)

// 事务参与者的服务地址
const SagaBarrierBusiApi = "/api/busi_saga_barrier"

var SagaBarrierBusi = fmt.Sprintf("http://localhost:%d%s", SagaBarrierBusiPort, SagaBarrierBusiApi)

func SagaBarrierMain() {
	go SagaBarrierStartSvr()
	SagaBarrierFireRequest()
	time.Sleep(1000 * time.Second)
}

func SagaBarrierStartSvr() {
	logrus.Printf("saga barrier examples starting")
	app := common.GetGinApp()
	SagaBarrierAddRoute(app)
	app.Run(fmt.Sprintf(":%d", SagaBarrierBusiPort))
}

func SagaBarrierFireRequest() {
	logrus.Printf("a busi transaction begin")
	req := &TransReq{
		Amount:         30,
		TransInResult:  "SUCCESS",
		TransOutResult: "SUCCESS",
	}
	saga := dtm.SagaNew(DtmServer).
		Add(SagaBarrierBusi+"/TransOut", SagaBarrierBusi+"/TransOutCompensate", req).
		Add(SagaBarrierBusi+"/TransIn", SagaBarrierBusi+"/TransInCompensate", req)
	logrus.Printf("busi trans commit")
	err := saga.Commit()
	e2p(err)
}

// api

func SagaBarrierAddRoute(app *gin.Engine) {
	app.POST(SagaBarrierBusiApi+"/TransIn", common.WrapHandler(sagaBarrierTransIn))
	app.POST(SagaBarrierBusiApi+"/TransInCompensate", common.WrapHandler(sagaBarrierTransInCompensate))
	app.POST(SagaBarrierBusiApi+"/TransOut", common.WrapHandler(sagaBarrierTransOut))
	app.POST(SagaBarrierBusiApi+"/TransOutCompensate", common.WrapHandler(sagaBarrierTransOutCompensate))
	logrus.Printf("examples listening at %d", SagaBarrierBusiPort)
}

var SagaBarrierTransInResult = ""
var SagaBarrierTransOutResult = ""
var SagaBarrierTransInCompensateResult = ""
var SagaBarrierTransOutCompensateResult = ""

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := reqFrom(c)
	res := common.OrString(SagaBarrierTransInResult, req.TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := reqFrom(c)
	res := common.OrString(SagaBarrierTransInCompensateResult, "SUCCESS")
	logrus.Printf("%s TransInCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	lid := c.Query("lid")
	req := reqFrom(c)
	return dtm.ThroughBarrierCall(dbGet().ToSqlDB(), "saga", gid, lid, "action", func(sdb *sql.DB) (interface{}, error) {
		db := common.SqlDB2DB(sdb)
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", c.Query("user_id")).
			Update("balance", gorm.Expr("balance - ?", req.Amount))
		return nil, dbr.Error
	})

	// res := common.OrString(SagaBarrierTransOutResult, req.TransOutResult, "SUCCESS")
	// logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	// return M{"result": res}, nil
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := reqFrom(c)
	res := common.OrString(SagaBarrierTransOutCompensateResult, "SUCCESS")
	logrus.Printf("%s TransOutCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

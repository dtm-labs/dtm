package examples

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm"
)

// 事务参与者的服务地址
const SagaBarrierBusiApi = "/api/busi_saga_barrier"

var SagaBarrierBusi = fmt.Sprintf("http://localhost:%d%s", SagaBarrierBusiPort, SagaBarrierBusiApi)

func SagaBarrierMainStart() {
	SagaBarrierStartSvr()
	SagaBarrierFireRequest()
}

func SagaBarrierStartSvr() {
	logrus.Printf("saga barrier examples starting")
	app := common.GetGinApp()
	SagaBarrierAddRoute(app)
	go app.Run(fmt.Sprintf(":%d", SagaBarrierBusiPort))
	time.Sleep(100 * time.Millisecond)
}

func SagaBarrierFireRequest() {
	logrus.Printf("a busi transaction begin")
	req := &TransReq{Amount: 30}
	saga := dtmcli.NewSaga(DtmServer).
		Add(SagaBarrierBusi+"/TransOut", SagaBarrierBusi+"/TransOutCompensate", req).
		Add(SagaBarrierBusi+"/TransIn", SagaBarrierBusi+"/TransInCompensate", req)
	logrus.Printf("busi trans submit")
	err := saga.Submit()
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

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		db := common.SqlDB2DB(sdb)
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", 1).
			Update("balance", gorm.Expr("balance + ?", req.Amount))
		return "SUCCESS", dbr.Error
	})
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		db := common.SqlDB2DB(sdb)
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", 1).
			Update("balance", gorm.Expr("balance - ?", req.Amount))
		return "SUCCESS", dbr.Error
	})
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		db := common.SqlDB2DB(sdb)
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", 2).
			Update("balance", gorm.Expr("balance - ?", req.Amount))
		return "SUCCESS", dbr.Error
	})
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		db := common.SqlDB2DB(sdb)
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", 2).
			Update("balance", gorm.Expr("balance + ?", req.Amount))
		return "SUCCESS", dbr.Error
	})
}

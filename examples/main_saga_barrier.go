package examples

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm"
)

// SagaBarrierFireRequest 1
func SagaBarrierFireRequest() {
	logrus.Printf("a busi transaction begin")
	req := &TransReq{Amount: 30}
	saga := dtmcli.NewSaga(DtmServer).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
	logrus.Printf("busi trans submit")
	err := saga.Submit()
	e2p(err)
}

// SagaBarrierAddRoute 1
func SagaBarrierAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/SagaBTransIn", common.WrapHandler(sagaBarrierTransIn))
	app.POST(BusiAPI+"/SagaBTransInCompensate", common.WrapHandler(sagaBarrierTransInCompensate))
	app.POST(BusiAPI+"/SagaBTransOut", common.WrapHandler(sagaBarrierTransOut))
	app.POST(BusiAPI+"/SagaBTransOutCompensate", common.WrapHandler(sagaBarrierTransOutCompensate))
	logrus.Printf("examples listening at %d", BusiPort)
}

func sagaBarrierAdjustBalance(sdb *sql.DB, uid int, amount int) (interface{}, error) {
	db := common.SQLDB2DB(sdb)
	dbr := db.Model(&UserAccount{}).Where("user_id = ?", 1).Update("balance", gorm.Expr("balance + ?", amount))
	return "SUCCESS", dbr.Error

}

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 1, reqFrom(c).Amount)
	})
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 1, -reqFrom(c).Amount)
	})
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 2, -reqFrom(c).Amount)
	})
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 2, reqFrom(c).Amount)
	})
}

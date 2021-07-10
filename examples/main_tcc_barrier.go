package examples

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// 事务参与者的服务地址
const TccBarrierBusiApi = "/api/busi_saga_barrier"

var TccBarrierBusi = fmt.Sprintf("http://localhost:%d%s", TccBarrierBusiPort, TccBarrierBusiApi)

func TccBarrierMainStart() {
	TccBarrierStartSvr()
	TccBarrierFireRequest()
}

func TccBarrierStartSvr() {
	logrus.Printf("saga barrier examples starting")
	app := common.GetGinApp()
	TccBarrierAddRoute(app)
	go app.Run(fmt.Sprintf(":%d", TccBarrierBusiPort))
	time.Sleep(100 * time.Millisecond)
}

func TccBarrierFireRequest() {
	logrus.Printf("tcc transaction begin")
	_, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&TransReq{Amount: 30}, TccBarrierBusi+"/TransOutTry", TccBarrierBusi+"/TransOutConfirm", TccBarrierBusi+"/TransOutRevert")
		if rerr != nil {
			return
		}
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&TransReq{Amount: 30}, TccBarrierBusi+"/TransInTry", TccBarrierBusi+"/TransInConfirm", TccBarrierBusi+"/TransInRevert")
		if rerr != nil {
			return
		}
		if res2.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res2.StatusCode())
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	e2p(err)
}

// api

func TccBarrierAddRoute(app *gin.Engine) {
	app.POST(TccBarrierBusiApi+"/TransInTry", common.WrapHandler(tccBarrierTransInTry))
	app.POST(TccBarrierBusiApi+"/TransInConfirm", common.WrapHandler(tccBarrierTransInConfirm))
	app.POST(TccBarrierBusiApi+"/TransInCancel", common.WrapHandler(tccBarrierTransInCancel))
	app.POST(TccBarrierBusiApi+"/TransOutTry", common.WrapHandler(tccBarrierTransOutTry))
	app.POST(TccBarrierBusiApi+"/TransOutConfirm", common.WrapHandler(tccBarrierTransOutConfirm))
	app.POST(TccBarrierBusiApi+"/TransOutCancel", common.WrapHandler(tccBarrierTransOutCancel))
	logrus.Printf("examples listening at %d", TccBarrierBusiPort)
}

const transInUid = 1
const transOutUid = 2

func adjustTrading(sdb *sql.DB, uid int, amount int) (interface{}, error) {
	db := common.SqlDB2DB(sdb)
	dbr := db.Exec("update dtm_busi.user_account_trading t join dtm_busi.user_account a on t.user_id=a.user_id and t.user_id=? set t.trading_balance=t.trading_balance + ? where a.balance + t.trading_balance + ? >= 0", uid, amount, amount)
	if dbr.Error == nil && dbr.RowsAffected == 0 {
		return nil, fmt.Errorf("update error, maybe balance not enough")
	}
	return "SUCCESS", nil
}

func adjustBalance(sdb *sql.DB, uid int, amount int) (interface{}, error) {
	db := common.SqlDB2DB(sdb)
	dbr := db.Exec("update dtm_busi.user_account_trading t join dtm_busi.user_account a on t.user_id=a.user_id and t.user_id=? set t.trading_balance=t.trading_balance + ?", uid, -amount, -amount)
	if dbr.Error == nil && dbr.RowsAffected == 1 {
		dbr = db.Exec("update dtm_busi.user_account set balance=balance+? where user_id=?", amount, uid)
	}
	if dbr.Error != nil {
		return nil, dbr.Error
	}
	if dbr.RowsAffected == 0 {
		return nil, fmt.Errorf("update trading error")
	}
	return "SUCCESS", nil
}

// TCC下，转入
func tccBarrierTransInTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transInUid, req.Amount)
	})
}

func tccBarrierTransInConfirm(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustBalance(sdb, transInUid, req.Amount)
	})
}

func tccBarrierTransInCancel(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transInUid, -req.Amount)
	})
}

func tccBarrierTransOutTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transOutUid, -req.Amount)
	})
}

func tccBarrierTransOutConfirm(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustBalance(sdb, transOutUid, -req.Amount)
	})
}

func tccBarrierTransOutCancel(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	return dtmcli.ThroughBarrierCall(dbGet().ToSqlDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transOutUid, req.Amount)
	})
}

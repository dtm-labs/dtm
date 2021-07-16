package examples

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func TccBarrierFireRequest() {
	logrus.Printf("tcc transaction begin")
	_, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutRevert")
		if rerr != nil {
			return
		}
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInRevert")
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
	app.POST(BusiApi+"/TccBTransInTry", common.WrapHandler(tccBarrierTransInTry))
	app.POST(BusiApi+"/TccBTransInConfirm", common.WrapHandler(tccBarrierTransInConfirm))
	app.POST(BusiApi+"/TccBTransInCancel", common.WrapHandler(tccBarrierTransInCancel))
	app.POST(BusiApi+"/TccBTransOutTry", common.WrapHandler(tccBarrierTransOutTry))
	app.POST(BusiApi+"/TccBTransOutConfirm", common.WrapHandler(tccBarrierTransOutConfirm))
	app.POST(BusiApi+"/TccBTransOutCancel", common.WrapHandler(tccBarrierTransOutCancel))
	logrus.Printf("examples listening at %d", BusiPort)
}

const transInUid = 1
const transOutUid = 2

func adjustTrading(sdb *sql.DB, uid int, amount int) (interface{}, error) {
	db := common.SQLDB2DB(sdb)
	dbr := db.Exec("update dtm_busi.user_account_trading t join dtm_busi.user_account a on t.user_id=a.user_id and t.user_id=? set t.trading_balance=t.trading_balance + ? where a.balance + t.trading_balance + ? >= 0", uid, amount, amount)
	if dbr.Error == nil && dbr.RowsAffected == 0 {
		return nil, fmt.Errorf("update error, maybe balance not enough")
	}
	return "SUCCESS", nil
}

func adjustBalance(sdb *sql.DB, uid int, amount int) (interface{}, error) {
	db := common.SQLDB2DB(sdb)
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
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transInUid, reqFrom(c).Amount)
	})
}

func tccBarrierTransInConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustBalance(sdb, transInUid, reqFrom(c).Amount)
	})
}

func tccBarrierTransInCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transInUid, -reqFrom(c).Amount)
	})
}

func tccBarrierTransOutTry(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transOutUid, -reqFrom(c).Amount)
	})
}

func tccBarrierTransOutConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustBalance(sdb, transOutUid, -reqFrom(c).Amount)
	})
}

func tccBarrierTransOutCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.TransInfoFromReq(c), func(sdb *sql.DB) (interface{}, error) {
		return adjustTrading(sdb, transOutUid, reqFrom(c).Amount)
	})
}

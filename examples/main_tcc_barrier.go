package examples

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// TccBarrierFireRequest 1
func TccBarrierFireRequest() string {
	logrus.Printf("tcc transaction begin")
	gid := dtmcli.MustGenGid(DtmServer)
	ret, err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	dtmcli.PanicIfFailure(ret, err)
	return gid
}

// TccBarrierAddRoute 1
func TccBarrierAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/TccBTransInTry", common.WrapHandler(tccBarrierTransInTry))
	app.POST(BusiAPI+"/TccBTransInConfirm", common.WrapHandler(tccBarrierTransInConfirm))
	app.POST(BusiAPI+"/TccBTransInCancel", common.WrapHandler(tccBarrierTransInCancel))
	app.POST(BusiAPI+"/TccBTransOutTry", common.WrapHandler(tccBarrierTransOutTry))
	app.POST(BusiAPI+"/TccBTransOutConfirm", common.WrapHandler(tccBarrierTransOutConfirm))
	app.POST(BusiAPI+"/TccBTransOutCancel", common.WrapHandler(TccBarrierTransOutCancel))
	logrus.Printf("examples listening at %d", BusiPort)
}

const transInUID = 1
const transOutUID = 2

func adjustTrading(sdb *sql.Tx, uid int, amount int) (interface{}, error) {
	affected, err := common.StxExec(sdb, "update dtm_busi.user_account_trading set trading_balance=trading_balance + ? where user_id=? and trading_balance + ? + (select balance from dtm_busi.user_account where id=?) >= 0", amount, uid, amount, uid)
	if err == nil && affected == 0 {
		return nil, fmt.Errorf("update error, maybe balance not enough")
	}
	return common.MS{"dtm_server": "SUCCESS"}, nil
}

func adjustBalance(sdb *sql.Tx, uid int, amount int) (interface{}, error) {
	affected, err := common.StxExec(sdb, "update dtm_busi.user_account_trading set trading_balance = trading_balance + ? where user_id=?;", -amount, uid)
	if err == nil && affected == 1 {
		affected, err = common.StxExec(sdb, "update dtm_busi.user_account set balance=balance+? where user_id=?", amount, uid)
	}
	if err == nil && affected == 0 {
		return nil, fmt.Errorf("update 0 rows")
	}
	return common.MS{"dtm_result": "SUCCESS"}, err
}

// TCC下，转入
func tccBarrierTransInTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c) // 去重构一下，改成可以重复使用的输入
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustTrading(sdb, transInUID, req.Amount)
	})
}

func tccBarrierTransInConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustBalance(sdb, transInUID, reqFrom(c).Amount)
	})
}

func tccBarrierTransInCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustTrading(sdb, transInUID, -reqFrom(c).Amount)
	})
}

func tccBarrierTransOutTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustTrading(sdb, transOutUID, -req.Amount)
	})
}

func tccBarrierTransOutConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustBalance(sdb, transOutUID, -reqFrom(c).Amount)
	})
}

// TccBarrierTransOutCancel will be use in test
func TccBarrierTransOutCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.ThroughBarrierCall(dbGet().ToSQLDB(), dtmcli.MustGetTrans(c), func(sdb *sql.Tx) (interface{}, error) {
		return adjustTrading(sdb, transOutUID, reqFrom(c).Amount)
	})
}

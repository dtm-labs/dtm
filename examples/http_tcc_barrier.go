package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func init() {
	setupFuncs["TccBarrierSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/TccBTransInTry", common.WrapHandler(tccBarrierTransInTry))
		app.POST(BusiAPI+"/TccBTransInConfirm", common.WrapHandler(tccBarrierTransInConfirm))
		app.POST(BusiAPI+"/TccBTransInCancel", common.WrapHandler(tccBarrierTransInCancel))
		app.POST(BusiAPI+"/TccBTransOutTry", common.WrapHandler(tccBarrierTransOutTry))
		app.POST(BusiAPI+"/TccBTransOutConfirm", common.WrapHandler(tccBarrierTransOutConfirm))
		app.POST(BusiAPI+"/TccBTransOutCancel", common.WrapHandler(TccBarrierTransOutCancel))
	}
	addSample("tcc_barrier", func() string {
		dtmcli.Logf("tcc transaction begin")
		gid := dtmcli.MustGenGid(DtmServer)
		err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransOutTry",
				Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		})
		dtmcli.FatalIfError(err)
		return gid
	})
}

const transInUID = 1
const transOutUID = 2

func adjustTrading(db dtmcli.DB, uid int, amount int) error {
	affected, err := dtmcli.DBExec(db, `update dtm_busi.user_account set trading_balance=trading_balance+?
		where user_id=? and trading_balance + ? + balance >= 0`, amount, uid, amount)
	if err == nil && affected == 0 {
		return fmt.Errorf("update error, maybe balance not enough")
	}
	return err
}

func adjustBalance(db dtmcli.DB, uid int, amount int) error {
	affected, err := dtmcli.DBExec(db, `update dtm_busi.user_account set trading_balance=trading_balance-?,
	  balance=balance+? where user_id=?`, amount, amount, uid)
	if err == nil && affected == 0 {
		return fmt.Errorf("update user_account 0 rows")
	}
	return err
}

// TCC下，转入
func tccBarrierTransInTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c) // 去重构一下，改成可以重复使用的输入
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustTrading(db, transInUID, req.Amount)
	})
}

func tccBarrierTransInConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustBalance(db, transInUID, reqFrom(c).Amount)
	})
}

func tccBarrierTransInCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustTrading(db, transInUID, -reqFrom(c).Amount)
	})
}

func tccBarrierTransOutTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustTrading(db, transOutUID, -req.Amount)
	})
}

func tccBarrierTransOutConfirm(c *gin.Context) (interface{}, error) {
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustBalance(db, transOutUID, -reqFrom(c).Amount)
	})
}

// TccBarrierTransOutCancel will be use in test
func TccBarrierTransOutCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(db dtmcli.DB) error {
		return adjustTrading(db, transOutUID, reqFrom(c).Amount)
	})
}

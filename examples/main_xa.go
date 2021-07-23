package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// XaClient XA client connection
var XaClient *dtmcli.XaClient = nil

// UserAccount busi model
type UserAccount struct {
	common.ModelBase
	UserID  int
	Balance string
}

// TableName gorm table name
func (u *UserAccount) TableName() string { return "user_account" }

// UserAccountTrading freeze user account table
type UserAccountTrading struct {
	common.ModelBase
	UserID         int
	TradingBalance string
}

// TableName gorm table name
func (u *UserAccountTrading) TableName() string { return "user_account_trading" }

func dbGet() *common.DB {
	return common.DbGet(config.Mysql)
}

// XaFireRequest 1
func XaFireRequest() string {
	gid := dtmcli.GenGid(DtmServer)
	err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (rerr error) {
		defer common.P2E(&rerr)
		req := GenTransReq(30, false, false)
		resp, err := xa.CallBranch(req, Busi+"/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = xa.CallBranch(req, Busi+"/TransInXa")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	e2p(err)
	return gid
}

// XaSetup 1
func XaSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInXa", common.WrapHandler(xaTransIn))
	app.POST(BusiAPI+"/TransOutXa", common.WrapHandler(xaTransOut))
	config.Mysql["database"] = "dtm_busi"
	XaClient = dtmcli.NewXaClient(DtmServer, config.Mysql, app, Busi+"/xa")
}

func xaTransIn(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c, func(db *common.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransInResult != "SUCCESS" {
			return fmt.Errorf("tranIn failed")
		}
		dbr := db.Exec("update user_account set balance=balance+? where user_id=?", req.Amount, 2)
		return dbr.Error
	})
	e2p(err)
	return M{"dtm_result": "SUCCESS"}, nil
}

func xaTransOut(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c, func(db *common.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransOutResult != "SUCCESS" {
			return fmt.Errorf("tranOut failed")
		}
		logrus.Printf("before updating balance")
		dbr := db.Exec("update user_account set balance=balance-? where user_id=?", req.Amount, 1)
		logrus.Printf("after updating balance")
		return dbr.Error
	})
	e2p(err)
	return M{"dtm_result": "SUCCESS"}, nil
}

// ResetXaData 1
func ResetXaData() {
	db := dbGet()
	db.Must().Exec("truncate user_account")
	db.Must().Exec("insert into user_account (user_id, balance) values (1, 10000), (2, 10000)")
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}

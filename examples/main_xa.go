package examples

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
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

// XaSetup 挂载http的api，创建XaClient
func XaSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInXa", common.WrapHandler(xaTransIn))
	app.POST(BusiAPI+"/TransOutXa", common.WrapHandler(xaTransOut))
	config.Mysql["database"] = "dtm_busi"
	var err error
	XaClient, err = dtmcli.NewXaClient(DtmServer, config.Mysql, app, Busi+"/xa")
	e2p(err)
}

// XaFireRequest 注册全局XA事务，调用XA的分支
func XaFireRequest() string {
	gid := dtmcli.MustGenGid(DtmServer)
	err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (rerr error) {
		defer common.P2E(&rerr)
		req := &TransReq{Amount: 30}
		resp, err := xa.CallBranch(req, Busi+"/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = xa.CallBranch(req, Busi+"/TransInXa")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	e2p(err)
	return gid
}

func xaTransIn(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c, func(db *common.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransInResult != "SUCCESS" {
			return fmt.Errorf("tranIn FAILURE")
		}
		dbr := db.Exec("update user_account set balance=balance+? where user_id=?", req.Amount, 2)
		return dbr.Error
	})
	if err != nil && strings.Contains(err.Error(), "FAILURE") {
		return M{"dtm_result": "FAILURE"}, nil
	}
	e2p(err)
	return M{"dtm_result": "SUCCESS"}, nil
}

func xaTransOut(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c, func(db *common.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransOutResult != "SUCCESS" {
			return fmt.Errorf("tranOut failed")
		}
		dbr := db.Exec("update user_account set balance=balance-? where user_id=?", req.Amount, 1)
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

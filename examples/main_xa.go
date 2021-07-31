package examples

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// XaClient XA client connection
var XaClient *dtmcli.XaClient = nil

func dbGet() *common.DB {
	return common.DbGet(config.DB)
}

func sdbGet() *sql.DB {
	return common.SdbGet(config.DB)
}

// XaSetup 挂载http的api，创建XaClient
func XaSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInXa", common.WrapHandler(xaTransIn))
	app.POST(BusiAPI+"/TransOutXa", common.WrapHandler(xaTransOut))
	var err error
	XaClient, err = dtmcli.NewXaClient(DtmServer, config.DB, app, Busi+"/xa")
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
	err := XaClient.XaLocalTransaction(c, func(db *sql.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransInResult == "FAILURE" {
			return fmt.Errorf("tranIn FAILURE")
		}
		_, rerr = common.SdbExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", req.Amount, 2)
		return
	})
	if err != nil && strings.Contains(err.Error(), "FAILURE") {
		return M{"dtm_result": "FAILURE"}, nil
	}
	e2p(err)
	return M{"dtm_result": "SUCCESS"}, nil
}

func xaTransOut(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c, func(db *sql.DB, xa *dtmcli.Xa) (rerr error) {
		req := reqFrom(c)
		if req.TransOutResult == "FAILURE" {
			return fmt.Errorf("tranOut failed")
		}
		_, rerr = common.SdbExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", req.Amount, 1)
		return
	})
	e2p(err)
	return M{"dtm_result": "SUCCESS"}, nil
}

// ResetXaData 1
func ResetXaData() {
	if config.DB["driver"] != "mysql" {
		return
	}
	db := dbGet()
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}

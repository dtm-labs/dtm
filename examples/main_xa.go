package examples

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// XaClient XA client connection
var XaClient *dtmcli.XaClient = nil

// XaSetup 挂载http的api，创建XaClient
func XaSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInXa", common.WrapHandler(xaTransIn))
	app.POST(BusiAPI+"/TransOutXa", common.WrapHandler(xaTransOut))
	var err error
	XaClient, err = dtmcli.NewXaClient(DtmServer, config.DB, Busi+"/xa", func(path string, xa *dtmcli.XaClient) {
		app.POST(path, common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			return xa.HandleCallback(c.Query("gid"), c.Query("branch_id"), c.Query("action"))
		}))
	})
	e2p(err)
}

// XaFireRequest 注册全局XA事务，调用XA的分支
func XaFireRequest() string {
	gid := dtmcli.MustGenGid(DtmServer)
	err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		resp, err := xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOutXa")
		if err != nil {
			return resp, err
		}
		return xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInXa")
	})
	e2p(err)
	return gid
}

func xaTransIn(c *gin.Context) (interface{}, error) {
	return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) (interface{}, error) {
		if reqFrom(c).TransInResult == "FAILURE" {
			return dtmcli.ResultFailure, nil
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", reqFrom(c).Amount, 2)
		return dtmcli.ResultSuccess, err
	})
}

func xaTransOut(c *gin.Context) (interface{}, error) {
	return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) (interface{}, error) {
		if reqFrom(c).TransOutResult == "FAILURE" {
			return dtmcli.ResultFailure, nil
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", reqFrom(c).Amount, 1)
		return dtmcli.ResultSuccess, err
	})
}

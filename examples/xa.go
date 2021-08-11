package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// XaClient XA client connection
var XaClient *dtmcli.XaClient = nil

func init() {
	setupFuncs["XaSetup"] = func(app *gin.Engine) {
		var err error
		XaClient, err = dtmcli.NewXaClient(DtmServer, config.DB, Busi+"/xa", func(path string, xa *dtmcli.XaClient) {
			app.POST(path, common.WrapHandler(func(c *gin.Context) (interface{}, error) {
				return xa.HandleCallback(c.Query("gid"), c.Query("branch_id"), c.Query("branch_type"))
			}))
		})
		dtmcli.FatalIfError(err)
	}
	addSample("xa", func() string {
		gid := dtmcli.MustGenGid(DtmServer)
		err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOutXa")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInXa")
		})
		dtmcli.FatalIfError(err)
		return gid
	})
}

package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func init() {
	setupFuncs["TccSetupSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/TransInTccParent", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			tcc, err := dtmcli.TccFromQuery(c.Request.URL.Query())
			dtmcli.FatalIfError(err)
			dtmcli.Logf("TransInTccParent ")
			return tcc.CallBranch(&TransReq{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		}))
	}
	addSample("tcc_nested", func() string {
		gid := dtmcli.MustGenGid(DtmServer)
		err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInTccParent", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		})
		dtmcli.FatalIfError(err)
		return gid
	})
	addSample("tcc", func() string {
		dtmcli.Logf("tcc simple transaction begin")
		gid := dtmcli.MustGenGid(DtmServer)
		err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		})
		dtmcli.FatalIfError(err)
		return gid
	})
}

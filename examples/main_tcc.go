package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// TccSetup 1
func TccSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInTccParent", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		tcc, err := dtmcli.TccFromReq(c)
		e2p(err)
		logrus.Printf("TransInTccParent ")
		return tcc.CallBranch(&TransReq{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	}))
}

// TccFireRequestNested 1
func TccFireRequestNested() string {
	gid := dtmcli.MustGenGid(DtmServer)
	ret, err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (interface{}, error) {
		resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInTccParent", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	dtmcli.PanicIfFailure(ret, err)
	return gid
}

// TccFireRequest 1
func TccFireRequest() string {
	logrus.Printf("tcc simple transaction begin")
	gid := dtmcli.MustGenGid(DtmServer)
	ret, err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (interface{}, error) {
		resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	dtmcli.PanicIfFailure(ret, err)
	return gid
}

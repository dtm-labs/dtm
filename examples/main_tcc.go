package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// TccSetup 1
func TccSetup(app *gin.Engine) {
	app.POST(BusiAPI+"/TransInTcc", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		tcc, err := dtmcli.TccFromReq(c)
		e2p(err)
		req := reqFrom(c)
		logrus.Printf("Trans in %d here, and Trans in another %d in call2 ", req.Amount/2, req.Amount/2)
		_, rerr := tcc.CallBranch(&TransReq{Amount: req.Amount / 2}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return M{"dtm_result": "SUCCESS"}, nil

	}))
}

// TccFireRequest 1
func TccFireRequest() string {
	logrus.Printf("tcc transaction begin")
	gid := dtmcli.MustGenGid(DtmServer)
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		res2, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInTcc", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	e2p(err)
	return gid
}

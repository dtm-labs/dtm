package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func TccSetup(app *gin.Engine) {
	app.POST(BusiApi+"/TransInTcc", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		tcc, err := dtmcli.TccFromReq(c)
		if err != nil {
			return nil, err
		}
		req := reqFrom(c)
		logrus.Printf("Trans in %d here, and Trans in another %d in call2 ", req.Amount/2, req.Amount/2)
		_, rerr := tcc.CallBranch(&TransReq{Amount: req.Amount / 2}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		if rerr != nil {
			return nil, rerr
		}

		return M{"result": "SUCCESS"}, nil

	}))
}

func TccFireRequest() {
	logrus.Printf("tcc transaction begin")
	_, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		if rerr != nil {
			return
		}
		res2, rerr := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInTcc", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		if rerr != nil {
			return
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	e2p(err)
}

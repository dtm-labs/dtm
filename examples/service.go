package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func AddRoute(app *gin.Engine) {
	app.POST(BusiApi+"/TransIn", common.WrapHandler(TransIn))
	app.POST(BusiApi+"/TransInCompensate", common.WrapHandler(TransInCompensate))
	app.POST(BusiApi+"/TransOut", common.WrapHandler(TransOut))
	app.POST(BusiApi+"/TransOutCompensate", common.WrapHandler(TransOutCompensate))
	app.GET(BusiApi+"/TransQuery", common.WrapHandler(TransQuery))
	logrus.Printf("examples istening at %d", BusiPort)
}

type M = map[string]interface{}

var TransInResult = ""
var TransOutResult = ""
var TransInCompensateResult = ""
var TransOutCompensateResult = ""
var TransQueryResult = ""

type TransReq struct {
	Amount         int  `json:"amount"`
	TransInFailed  bool `json:"transInFailed"`
	TransOutFailed bool `json:"transOutFailed"`
}

func TransIn(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}
	if req.TransInFailed {
		logrus.Printf("%s TransIn %v failed", gid, req)
		return M{"result": "FAIL"}, nil
	}
	res := common.OrString(TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransInCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}
	res := common.OrString(TransInCompensateResult, "SUCCESS")
	logrus.Printf("%s TransInCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransOut(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}
	if req.TransOutFailed {
		logrus.Printf("%s TransOut %v failed", gid, req)
		return M{"result": "FAIL"}, nil
	}
	res := common.OrString(TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransOutCompensate(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}
	res := common.OrString(TransOutCompensateResult, "SUCCESS")
	logrus.Printf("%s TransOutCompensate: %v result: %s", gid, req, res)
	return M{"result": res}, nil
}

func TransQuery(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	res := common.OrString(TransQueryResult, "SUCCESS")
	return M{"result": res}, nil
}

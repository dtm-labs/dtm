package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtm"
)

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

func TransIn(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	if req.TransInFailed {
		logrus.Printf("%s TransIn %v failed", req)
		c.Error(fmt.Errorf("TransIn failed for gid: %s", gid))
		return
	}
	res := common.OrString(TransInResult, "SUCCESS")
	logrus.Printf("%s TransIn: %v result: %s", gid, req, res)
	c.JSON(200, M{"result": res})
}

func TransInCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	res := common.OrString(TransInCompensateResult, "SUCCESS")
	logrus.Printf("%s TransInCompensate: %v result: %s", gid, req, res)
	c.JSON(200, M{"result": res})
}

func TransOut(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	if req.TransOutFailed {
		logrus.Printf("%s TransOut %v failed", gid, req)
		c.JSON(500, M{"result": "FAIL"})
		return
	}
	res := common.OrString(TransOutResult, "SUCCESS")
	logrus.Printf("%s TransOut: %v result: %s", gid, req, res)
	c.JSON(200, M{"result": res})
}

func TransOutCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	res := common.OrString(TransOutCompensateResult, "SUCCESS")
	logrus.Printf("%s TransOutCompensate: %v result: %s", gid, req, res)
	c.JSON(200, M{"result": res})
}

func TransQuery(c *gin.Context) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	res := common.OrString(TransQueryResult, "SUCCESS")
	c.JSON(200, M{"result": res})
}

func trans(req *TransReq) {
	// gid := common.GenGid()
	gid := "4eHhkCxVsQ1"
	logrus.Printf("busi transaction begin: %s", gid)
	saga := dtm.SagaNew(TcServer, gid, Busi+"/TransQuery")

	saga.Add(Busi+"/TransIn", Busi+"/TransInCompensate", M{
		"amount":         req.Amount,
		"transInFailed":  req.TransInFailed,
		"transOutFailed": req.TransOutFailed,
	})
	saga.Add(Busi+"/TransOut", Busi+"/TransOutCompensate", M{
		"amount":         req.Amount,
		"transInFailed":  req.TransInFailed,
		"transOutFailed": req.TransOutFailed,
	})
	saga.Prepare()
	logrus.Printf("busi trans commit")
	saga.Commit()
}

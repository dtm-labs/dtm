package examples

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtm"
)

type M = map[string]interface{}
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
	logrus.Printf("%s TransIn: %v", gid, req)
	if req.TransInFailed {
		logrus.Printf("%s TransIn %v failed", req)
		c.Error(fmt.Errorf("TransIn failed for gid: %s", gid))
		return
	}
	c.JSON(200, M{"result": "SUCCESS"})
}

func TransInCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransInCompensate: %v", gid, req)
	c.JSON(200, M{"result": "SUCCESS"})
}

func TransOut(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransOut: %v", gid, req)
	if req.TransOutFailed {
		logrus.Printf("%s TransOut %v failed", gid, req)
		c.JSON(500, M{"result": "FAIL"})
		return
	}
	c.JSON(200, M{"result": "SUCCESS"})
}

func TransOutCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransOutCompensate: %v", gid, req)
	c.JSON(200, M{"result": "SUCCESS"})
}

func TransQuery(c *gin.Context) {
	gid := c.Query("gid")
	logrus.Printf("%s TransQuery", gid)
	if strings.Contains(gid, "cancel") {
		c.JSON(200, M{"result": "FAIL"})
	} else if strings.Contains(gid, "pending") {
		c.JSON(200, M{"result": "PENDING"})
	} else {
		c.JSON(200, M{"result": "SUCCESS"})
	}
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

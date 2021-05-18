package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtm"
)

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
	c.JSON(200, gin.H{"result": "SUCCESS"})
}

func TransInCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransInCompensate: %v", gid, req)
	c.JSON(200, gin.H{"result": "SUCCESS"})
}

func TransOut(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	req2 := TransReq{}
	c.ShouldBindBodyWith(&req2, binding.JSON)
	c.ShouldBindBodyWith(&req, binding.JSON)
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransOut: %v", gid, req)
	if req.TransOutFailed {
		logrus.Printf("%s TransOut %v failed", gid, req)
		c.JSON(500, gin.H{"result": "FAIL"})
		return
	}
	c.JSON(200, gin.H{"result": "SUCCESS"})
}

func TransOutCompensate(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransOutCompensate: %v", gid, req)
	c.JSON(200, gin.H{"result": "SUCCESS"})
}

func TransQuery(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransQuery: %v", gid, req)
	c.JSON(200, gin.H{"result": "SUCCESS"})
}

func trans(req *TransReq) {
	// gid := common.GenGid()
	gid := "4eHhkCxVsQ1"
	logrus.Printf("busi transaction begin: %s", gid)
	saga := dtm.SagaNew(TcServer, gid)

	saga.Add(Busi+"/TransIn", Busi+"/TransInCompensate", gin.H{
		"amount":         req.Amount,
		"transInFailed":  req.TransInFailed,
		"transOutFailed": req.TransOutFailed,
	})
	saga.Add(Busi+"/TransOut", Busi+"/TransOutCompensate", gin.H{
		"amount":         req.Amount,
		"transInFailed":  req.TransInFailed,
		"transOutFailed": req.TransOutFailed,
	})
	saga.Prepare(Busi + "/TransQuery")
	logrus.Printf("busi trans commit")
	saga.Commit()
}

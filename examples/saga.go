package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtm"
)

type TransReq struct {
	amount         int
	transInFailed  bool
	transOutFailed bool
}

func TransIn(c *gin.Context) {
	gid := c.Query("gid")
	req := TransReq{}
	if err := c.BindJSON(&req); err != nil {
		return
	}
	logrus.Printf("%s TransIn: %v", gid, req)
	if req.transInFailed {
		logrus.Printf("%s TransIn %v failed", req)
		c.Error(fmt.Errorf("TransIn failed for gid: %s", gid))
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
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	saga := dtm.Saga{
		Server: TcServer,
		Gid:    gid,
	}

	saga.Add(Busi+"/TransIn", Busi+"/TransInCompensate", gin.H{
		"amount":         req.amount,
		"transInFailed":  req.transInFailed,
		"transOutFailed": req.transOutFailed,
	})
	saga.Prepare(Busi + "/TransQuery")
	logrus.Printf("busi trans commit")
	saga.Commit()
}

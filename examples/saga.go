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

func TansIn(c *gin.Context) {
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

func TansInCompensate(c *gin.Context) {
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

	saga.Add(Busi+"/TransIn", Busi+"/TransInCompensate", req)
	saga.Prepare(Busi + "TransQuery")
	logrus.Printf("busi trans commit")
	saga.Commit()
}

/*
export async function transQuery(ctx: ServiceContext) {
  let { gid } = ctx.query
  return { message: 'SUCCESS' }
  // return `gid: ${gid} ` + (Math.random() * 3 > 1 ? 'SUCCESS' : 'FAIL')
}

let host = "http://localhost:4005"

export async function startSagaTrans(ctx: ServiceContext) {
  await ctx.beginTransaction()
  let gid = await getGlobalTid()
  console.log(`order: ${gid} created`)
  let { transIn, transOut } = ctx.data
  let saga = new Saga(`${host}/api/core/saga-svr`, gid)
  saga.add(`${host}/api/core/dtrans/transIn`, `${host}/api/core/dtrans/transInCompensate`, { amount: 30, transIn, transOut })
  saga.add(`${host}/api/core/dtrans/transOut`, `${host}/api/core/dtrans/transOutCompensate`, { amount: 30, transIn, transOut })
  await saga.prepare(`${host}/api/core/dtrans/transQuery`)
  await ctx.trans.commit()
  await saga.commit()
}
*/

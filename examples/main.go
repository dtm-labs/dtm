package examples

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtm"
)

func FireRequest() {
	gid := common.GenGid()
	logrus.Printf("busi transaction begin: %s", gid)
	req := &TransReq{
		Amount:         30,
		TransInFailed:  false,
		TransOutFailed: false,
	}
	saga := dtm.SagaNew(DtmServer, gid, Busi+"/TransQuery")

	saga.Add(Busi+"/TransIn", Busi+"/TransInCompensate", req)
	saga.Add(Busi+"/TransOut", Busi+"/TransOutCompensate", req)
	saga.Prepare()
	logrus.Printf("busi trans commit")
	saga.Commit()
}

func StartSvr() {
	logrus.Printf("examples starting")
	app := common.GetGinApp()
	app.POST(BusiApi+"/TransIn", TransIn)
	app.POST(BusiApi+"/TransInCompensate", TransInCompensate)
	app.POST(BusiApi+"/TransOut", TransOut)
	app.POST(BusiApi+"/TransOutCompensate", TransOutCompensate)
	app.GET(BusiApi+"/TransQuery", TransQuery)
	logrus.Printf("examples istening at %d", BusiPort)
	app.Run(":8081")
}

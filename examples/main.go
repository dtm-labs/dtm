package examples

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtm"
)

func Main() {
	go StartSvr()
	FireRequest()
	time.Sleep(1000 * time.Second)
}

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
	err := saga.Prepare()
	common.PanicIfError(err)
	logrus.Printf("busi trans commit")
	err = saga.Commit()
	common.PanicIfError(err)
}

func StartSvr() {
	logrus.Printf("examples starting")
	app := common.GetGinApp()
	AddRoute(app)
	app.Run(":8081")
}

package examples

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func Main() {
	go StartSvr()
	trans(&TransReq{
		Amount:         30,
		TransInFailed:  false,
		TransOutFailed: true,
	})
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

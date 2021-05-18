package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Main() {
	logrus.Printf("examples starting")
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	app.POST(BusiApi+"/TransIn", TransIn)
	app.POST(BusiApi+"/TransInCompensate", TransInCompensate)
	app.POST(BusiApi+"/TransOut", TransOut)
	app.POST(BusiApi+"/TransOutCompensate", TransOutCompensate)
	app.POST(BusiApi+"/TransQuery", TransQuery)

	go app.Run(":8081")
	logrus.Printf("examples istening at %d", BusiPort)
	trans(&TransReq{
		Amount:         30,
		TransInFailed:  false,
		TransOutFailed: true,
	})
}

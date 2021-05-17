package dtmsvr

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Printf("start tc")
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	AddRoute(app)
	go ConsumeHalfMsg()
	go ConsumeMsg()
	app.Run()
}

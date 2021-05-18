package dtmsvr

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Main() {
	logrus.Printf("start dtmsvr")
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	AddRoute(app)
	// StartConsumePreparedMsg(1)
	StartConsumeCommitedMsg(1)
	logrus.Printf("dtmsvr listen at: 8080")
	go app.Run()
}

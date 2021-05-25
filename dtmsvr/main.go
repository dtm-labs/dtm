package dtmsvr

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func Main() {
	go StartSvr()
}

func StartSvr() {
	logrus.Printf("start dtmsvr")
	common.InitApp(&config)
	app := common.GetGinApp()
	AddRoute(app)
	logrus.Printf("dtmsvr listen at: 8080")
	app.Run(":8080")
}

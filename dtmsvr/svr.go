package dtmsvr

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func Main() {
	logrus.Printf("dtmsvr listen at: 8080")
	go StartSvr()
}

func StartSvr() {
	logrus.Printf("start dtmsvr")
	app := common.GetGinApp()
	AddRoute(app)
	logrus.Printf("dtmsvr listen at: 8080")
	app.Run()
}

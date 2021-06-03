package dtmsvr

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080

func Main() {
	go StartSvr()
}

func StartSvr() {
	logrus.Printf("start dtmsvr")
	common.InitApp(common.GetCurrentPath(), &config)
	app := common.GetGinApp()
	AddRoute(app)
	logrus.Printf("dtmsvr listen at: %d", dtmsvrPort)
	app.Run(fmt.Sprintf(":%d", dtmsvrPort))
}

func PopulateMysql() {
	common.InitApp(common.GetCurrentPath(), &config)
	examples.RunSqlScript(config.Mysql, common.GetCurrentPath()+"/dtmsvr.sql")
}

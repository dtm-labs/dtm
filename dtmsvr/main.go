package dtmsvr

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

func Main() {
	go StartSvr()
}

func StartSvr() {
	logrus.Printf("start dtmsvr")
	common.InitApp(common.GetCurrentPath(), &config)
	app := common.GetGinApp()
	AddRoute(app)
	logrus.Printf("dtmsvr listen at: 8080")
	app.Run(":8080")
}

func PopulateMysql() {
	common.InitApp(common.GetCurrentPath(), &config)
	examples.RunSqlScript(config.Mysql, common.GetCurrentPath()+"/dtmsvr.sql")
}

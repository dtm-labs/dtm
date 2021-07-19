package dtmsvr

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080

// MainStart main
func MainStart() {
	StartSvr()
	go CronExpiredTrans(-1)
}

// StartSvr StartSvr
func StartSvr() {
	logrus.Printf("start dtmsvr")
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
	app := common.GetGinApp()
	addRoute(app)
	logrus.Printf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))
	time.Sleep(100 * time.Millisecond)
}

// PopulateMysql setup mysql data
func PopulateMysql() {
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = ""
	examples.RunSQLScript(config.Mysql, common.GetCurrentCodeDir()+"/dtmsvr.sql")
}

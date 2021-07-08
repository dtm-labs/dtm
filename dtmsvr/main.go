package dtmsvr

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080

func MainStart() {
	StartSvr()
	go CronCommitted()
	go CronPrepared()
}

func StartSvr() {
	logrus.Printf("start dtmsvr")
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
	app := common.GetGinApp()
	AddRoute(app)
	logrus.Printf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))
	time.Sleep(100 * time.Millisecond)
}

func PopulateMysql() {
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = ""
	examples.RunSqlScript(config.Mysql, common.GetCurrentDir()+"/dtmsvr.sql")
}

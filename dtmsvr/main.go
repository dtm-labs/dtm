package dtmsvr

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080

// StartSvr StartSvr
func StartSvr() {
	logrus.Printf("start dtmsvr")
	app := common.GetGinApp()
	addRoute(app)
	logrus.Printf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))
	time.Sleep(100 * time.Millisecond)
}

// PopulateMysql setup mysql data
func PopulateMysql(skipDrop bool) {
	examples.RunSQLScript(config.Mysql, common.GetCurrentCodeDir()+"/dtmsvr.sql", skipDrop)
}

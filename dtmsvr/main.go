package dtmsvr

import (
	"fmt"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080

// StartSvr StartSvr
func StartSvr() {
	dtmcli.Logf("start dtmsvr")
	app := common.GetGinApp()
	addRoute(app)
	dtmcli.Logf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))
	time.Sleep(100 * time.Millisecond)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCurrentCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

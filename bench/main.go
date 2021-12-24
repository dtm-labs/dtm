package main

import (
	"fmt"
	"os"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/dtm/examples"
)

var hint = `To start the bench server, you need to specify the parameters:

Available commands:
    http              start bench server
`

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf(hint)
		return
	}
	logger.Debugf("starting dtm....")
	if os.Args[1] == "http" {
		fmt.Println("start bench server")
		common.MustLoadConfig()
		dtmcli.SetCurrentDBType(common.Config.ExamplesDB.Driver)
		registry.WaitStoreUp()
		dtmsvr.PopulateDB(true)
		examples.PopulateDB(true)
		dtmsvr.StartSvr()              // 启动dtmsvr的api服务
		go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询
		StartSvr()
		select {}
	} else {
		fmt.Printf(hint)
	}
}

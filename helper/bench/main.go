package main

import (
	"fmt"
	"os"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/dtm/helper/bench/svr"
	"github.com/dtm-labs/dtm/test/busi"
)

var usage = `bench is a bench test server for dtmf
usage:
    redis   prepare for redis bench test
    db      prepare for mysql|postgres bench test
		boltdb  prepare for boltdb bench test
`

func hintAndExit() {
	fmt.Print(usage)
	os.Exit(0)
}

var conf = &config.Config

func main() {
	if len(os.Args) <= 1 {
		hintAndExit()
	}
	logger.Infof("starting bench server")
	config.MustLoadConfig("")
	logger.InitLog(conf.LogLevel)
	registry.WaitStoreUp()
	dtmsvr.PopulateDB(false)
	if os.Args[1] == "db" {
		if busi.BusiConf.Driver == "mysql" {
			dtmcli.SetCurrentDBType(busi.BusiConf.Driver)
			svr.PrepareBenchDB()
		}
		busi.PopulateDB(false)
	} else if os.Args[1] == "redis" || os.Args[1] == "boltdb" {

	} else {
		hintAndExit()
	}
	dtmsvr.StartSvr()              // 启动dtmsvr的api服务
	go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询
	svr.StartSvr()                 // 启动bench服务
	select {}
}

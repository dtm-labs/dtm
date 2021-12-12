package main

import (
	"fmt"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		dtmimp.Logf("starting dtm....")
		if os.Args[1] == "http" {
			fmt.Println("start bench server")
			common.MustLoadConfig()
			dtmcli.SetCurrentDBType(common.DtmConfig.DB["driver"])
			common.WaitDBUp()
			dtmsvr.PopulateDB(true)
			examples.PopulateDB(true)
			dtmsvr.StartSvr()              // 启动dtmsvr的api服务
			go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询
			StartSvr()
			select {}
		}
	}
}

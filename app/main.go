package main

import (
	"os"
	"time"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

// M alias
type M = map[string]interface{}

func wait() {
	for {
		time.Sleep(10000 * time.Second)
	}
}

func main() {
	onlyServer := len(os.Args) > 1 && os.Args[1] == "dtmsvr"
	if !onlyServer { // 实际线上运行，只启动dtmsvr，不准备table相关的数据
		dtmsvr.PopulateDB(true)
	}
	dtmsvr.StartSvr()              // 启动dtmsvr的api服务
	go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询

	if onlyServer || len(os.Args) == 1 { // 没有参数，或者参数为dtmsvr，则不运行例子
		wait()
	}

	examples.PopulateDB(true)

	// quick_start 比较独立，单独作为一个例子运行，方便新人上手
	if len(os.Args) > 1 && (os.Args[1] == "quick_start" || os.Args[1] == "qs") {
		examples.QsStartSvr()
		examples.QsFireRequest()
		wait()
	}

	// 下面是各类的例子
	app := examples.BaseAppStartup()
	if os.Args[1] == "xa" { // 启动xa示例
		examples.XaSetup(app)
		examples.XaFireRequest()
	} else if os.Args[1] == "saga" { // 启动saga示例
		examples.SagaSetup(app)
		examples.SagaFireRequest()
	} else if os.Args[1] == "tcc" { // 启动tcc示例
		examples.TccSetup(app)
		examples.TccFireRequestNested()
	} else if os.Args[1] == "msg" { // 启动msg示例
		examples.MsgSetup(app)
		examples.MsgFireRequest()
	} else if os.Args[1] == "all" { // 运行所有示例
		examples.SagaSetup(app)
		examples.SagaWaitSetup(app)
		examples.TccSetup(app)
		examples.XaSetup(app)
		examples.MsgSetup(app)
		examples.SagaFireRequest()
		examples.TccFireRequestNested()
		examples.XaFireRequest()
		examples.MsgFireRequest()
	} else if os.Args[1] == "saga_barrier" {
		examples.SagaBarrierAddRoute(app)
		examples.SagaBarrierFireRequest()
	} else if os.Args[1] == "tcc_barrier" {
		examples.TccBarrierAddRoute(app)
		examples.TccBarrierFireRequest()
	} else {
		dtmcli.LogRedf("unknown arg: %s", os.Args[1])
	}
	wait()
}

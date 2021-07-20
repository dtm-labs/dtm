package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
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
	if len(os.Args) > 1 && os.Args[1] == "dtmsvr" { // 实际运行，只启动dtmsvr，不重新load数据
		dtmsvr.MainStart()
		wait()
	}
	// 下面都是运行示例，因此首先把服务器的数据重新准备好
	dtmsvr.PopulateMysql()
	dtmsvr.MainStart()
	if len(os.Args) == 1 { // 默认没有参数的情况下，准备好数据并启动dtmsvr即可
		wait()
	}
	// quick_start 比较独立，单独作为一个例子运行，方便新人上手
	if len(os.Args) > 1 && (os.Args[1] == "quick_start" || os.Args[1] == "qs") {
		examples.QsStartSvr()
		examples.QsFireRequest()
		wait()
	}

	// 下面是各类的例子
	examples.PopulateMysql()
	app := examples.BaseAppStartup()
	if os.Args[1] == "xa" { // 启动xa示例
		examples.XaSetup(app)
		examples.XaFireRequest()
	} else if os.Args[1] == "saga" { // 启动saga示例
		examples.SagaSetup(app)
		examples.SagaFireRequest()
	} else if os.Args[1] == "tcc" { // 启动tcc示例
		examples.TccSetup(app)
		examples.TccFireRequest()
	} else if os.Args[1] == "msg" { // 启动msg示例
		examples.MsgSetup(app)
		examples.MsgFireRequest()
	} else if os.Args[1] == "all" { // 运行所有示例
		examples.SagaSetup(app)
		examples.TccSetup(app)
		examples.XaSetup(app)
		examples.MsgSetup(app)
		examples.SagaFireRequest()
		examples.TccFireRequest()
		examples.XaFireRequest()
		examples.MsgFireRequest()
	} else if os.Args[1] == "saga_barrier" {
		examples.SagaBarrierAddRoute(app)
		examples.SagaBarrierFireRequest()
	} else if os.Args[1] == "tcc_barrier" {
		examples.TccBarrierAddRoute(app)
		examples.TccBarrierFireRequest()
	} else {
		logrus.Fatalf("unknown arg: %s", os.Args[1])
	}
	wait()
}

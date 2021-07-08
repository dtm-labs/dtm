package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

type M = map[string]interface{}

func wait() {
	time.Sleep(10000 * time.Second)
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "quick_start" || os.Args[1] == "qs") {
		dtmsvr.PopulateMysql()
		dtmsvr.MainStart()
		examples.StartMain()
		wait()
	}
	app := examples.BaseAppNew()
	examples.BaseAppSetup(app)
	if len(os.Args) == 1 || os.Args[1] == "saga" { // 默认情况下，展示saga例子
		dtmsvr.PopulateMysql()
		dtmsvr.MainStart()
		examples.SagaSetup(app)
		examples.BaseAppStart(app)
		examples.SagaFireRequest()
	} else if os.Args[1] == "xa" { // 启动xa示例
		dtmsvr.PopulateMysql()
		dtmsvr.MainStart()
		examples.PopulateMysql()
		examples.XaSetup(app)
		examples.BaseAppStart(app)
		examples.XaFireRequest()
	} else if os.Args[1] == "dtmsvr" { // 只启动dtmsvr
		go dtmsvr.MainStart()
	} else if os.Args[1] == "all" { // 运行所有示例
		dtmsvr.PopulateMysql()
		examples.PopulateMysql()
		dtmsvr.MainStart()
		examples.SagaSetup(app)
		examples.TccSetup(app)
		examples.XaSetup(app)
		examples.BaseAppStart(app)
		examples.SagaFireRequest()
		examples.TccFireRequest()
		examples.XaFireRequest()
	} else if os.Args[1] == "saga_barrier" {
		dtmsvr.PopulateMysql()
		dtmsvr.MainStart()
		examples.PopulateMysql()
		examples.SagaBarrierMainStart()
	} else {
		logrus.Fatalf("unknown arg: %s", os.Args[1])
	}
	wait()
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

// M alias
type M = map[string]interface{}

var (
	onlyServer          bool
	tutorial            string
	expectTutorialValue = false
)

func wait() {
	reload := make(chan int, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-reload:
		case sg := <-stop:
			if sg == syscall.SIGHUP {
			} else {
				dtmsvr.StopSvr()
				if len(tutorial) > 0 && expectTutorialValue {
					examples.StopExampleSvr()
				}
				return
			}
		}
	}
}

func init() {
	flag.BoolVar(&onlyServer, "bare", false, "only run dtm server, skip create table")
	flag.StringVar(&tutorial, "tutorial", "", "choose which example you want run, should be in [quick_start,qs,xa,saga,tcc,msg,all,saga_barrier,tcc_barrier]")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) > 1 {
		dtmcli.LogRedf("os.Args[1] value: %+v", os.Args[1])
		if os.Args[1] == "dtmsvr" {
			onlyServer = true
		} else {
			tutorial = os.Args[1]
		}
	}
	flag.Parse()
	dtmcli.LogRedf("bare value: %+v", onlyServer)
	dtmcli.LogRedf("tutorial value: %+v", tutorial)
	if !onlyServer { // 实际线上运行，只启动dtmsvr，不准备table相关的数据
		dtmsvr.PopulateDB(true)
	}
	dtmsvr.StartSvr(0)             // 启动dtmsvr的api服务
	go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询

	if len(tutorial) == 0 { // 没有参数，或者参数为dtmsvr，则不运行例子
		wait()
		return
	}

	examples.PopulateDB(true)
	expectTutorialValue = true
	// quick_start 比较独立，单独作为一个例子运行，方便新人上手
	if tutorial == "quick_start" || tutorial == "qs" {
		examples.QsStartSvr(0)
		examples.QsFireRequest()
		wait()
		return
	}
	// 下面是各类的例子
	app := examples.BaseAppStartup(0)
	switch tutorial {
	case "xa":
		examples.XaSetup(app)
		examples.XaFireRequest()
	case "saga":
		examples.SagaSetup(app)
		examples.SagaFireRequest()
	case "tcc":
		examples.TccSetup(app)
		examples.TccFireRequestNested()
	case "msg":
		examples.MsgSetup(app)
		examples.MsgFireRequest()
	case "all":
		examples.SagaSetup(app)
		examples.TccSetup(app)
		examples.XaSetup(app)
		examples.MsgSetup(app)
		examples.SagaFireRequest()
		examples.TccFireRequestNested()
		examples.XaFireRequest()
		examples.MsgFireRequest()
	case "saga_barrier":
		examples.SagaBarrierAddRoute(app)
		examples.SagaBarrierFireRequest()
	case "tcc_barrier":
		examples.TccBarrierAddRoute(app)
		examples.TccBarrierFireRequest()
	default:
		expectTutorialValue = false
		dtmcli.LogRedf("unknown tutorial value: %s", tutorial)
	}
	wait()
	return
}

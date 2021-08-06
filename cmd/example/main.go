package example

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func main() {
	reload := make(chan int, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	startServer()
	for {
		select {
		case <-reload:
		case sg := <-stop:
			if sg == syscall.SIGHUP {
				startServer()
			} else {
				dtmsvr.StopSvr()
				return
			}
		}
	}
}

func startServer() {
	if isDev {
		dtmsvr.PopulateDB(true)
	}
	dtmsvr.StartSvr(port)          // 启动dtmsvr的api服务
	go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询
	examples.PopulateDB(true)

	if tutorial == "quick_start" || tutorial == "qs" {
		examples.QsStartSvr(examplePort)
		examples.QsFireRequest()
	} else {
		app := examples.BaseAppStartup(examplePort)
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
			dtmcli.LogRedf("unknown arg: %s", os.Args[1])
		}
	}
}

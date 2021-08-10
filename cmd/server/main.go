package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/yedf/dtm/dtmsvr"
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
}

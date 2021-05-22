package main

import (
	"os"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

type M = map[string]interface{}

func main() {
	cmd := common.If(len(os.Args) > 1, os.Args[1], "").(string)
	dtmsvr.LoadConfig()
	if cmd == "" { // 所有服务都启动
		go dtmsvr.StartSvr()
		go examples.StartSvr()
	} else if cmd == "dtmsvr" {
		go dtmsvr.StartSvr()
	}
	time.Sleep(1000 * 1000 * 1000 * 1000)
}

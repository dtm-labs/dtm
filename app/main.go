package main

import (
	"os"
	"time"

	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

type M = map[string]interface{}

func main() {
	if len(os.Args) == 1 { // 所有服务都启动
		go dtmsvr.StartSvr()
		go examples.SagaStartSvr()
	} else if len(os.Args) > 1 && os.Args[1] == "dtmsvr" {
		go dtmsvr.StartSvr()
	}
	for {
		time.Sleep(1000 * time.Second)
	}
}

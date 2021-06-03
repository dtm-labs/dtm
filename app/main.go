package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

type M = map[string]interface{}

func main() {
	if len(os.Args) == 1 { // 默认情况下，展示saga例子
		dtmsvr.PopulateMysql()
		go dtmsvr.StartSvr()
		go examples.SagaStartSvr()
		time.Sleep(100 * time.Millisecond)
		examples.SagaFireRequest()
	} else if os.Args[1] == "dtmsvr" {
		go dtmsvr.StartSvr()
	} else if os.Args[1] == "all" {
		dtmsvr.PopulateMysql()
		examples.PopulateMysql()
		go dtmsvr.StartSvr()
		go examples.SagaStartSvr()
		go examples.TccStartSvr()
		go examples.XaStartSvr()
	} else {
		logrus.Fatalf("unknown arg: %s", os.Args[1])
	}
	for {
		time.Sleep(1000 * time.Second)
	}
}

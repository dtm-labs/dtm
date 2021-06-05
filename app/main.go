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
		go dtmsvr.Main()
		go examples.SagaStartSvr()
		time.Sleep(100 * time.Millisecond)
		examples.SagaFireRequest()
	} else if os.Args[1] == "dtmsvr" { // 只启动dtmsvr
		go dtmsvr.StartSvr()
	} else if os.Args[1] == "all" { // 运行所有示例
		dtmsvr.PopulateMysql()
		examples.PopulateMysql()
		go dtmsvr.Main()
		go examples.SagaStartSvr()
		go examples.TccStartSvr()
		go examples.XaStartSvr()
		time.Sleep(100 * time.Millisecond)
		examples.SagaFireRequest()
		examples.TccFireRequest()
		examples.XaFireRequest()
	} else {
		logrus.Fatalf("unknown arg: %s", os.Args[1])
	}
	for {
		time.Sleep(1000 * time.Second)
	}
}

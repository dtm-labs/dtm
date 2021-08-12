package main

import (
	"fmt"
	"os"
	"strings"
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
	if len(os.Args) == 1 {
		for _, ln := range []string{
			"dtm is a lightweight distributed transaction manager.",
			"usage:",
			"dtm [command]",
			"",
			"Available Commands:",
			"dtmsvr            run dtm as a server. ",
			"",
			"quick_start       run quick start example. dtm will create all needed table",
			"qs                same as quick_start",
		} {
			fmt.Print(ln + "\n")
		}
		for name := range examples.Samples {
			fmt.Printf("%-18srun a sample includes %s\n", name, strings.Replace(name, "_", " ", 100))
		}
		return
	}
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
	examples.GrpcStartup()
	examples.BaseAppStartup()

	sample := examples.Samples[os.Args[1]]
	dtmcli.LogIfFatalf(sample == nil, "no sample name for %s", os.Args[1])
	sample.Action()
	wait()
}

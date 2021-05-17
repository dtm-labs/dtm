package main

import (
	"time"

	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func main() {
	dtmsvr.LoadConfig()
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	// dtmsvr.LoadConfig()
	// rb := dtmsvr.RabbitmqNew(&dtmsvr.ServerConfig.Rabbitmq)
	// err := rb.SendAndConfirm(dtmsvr.RabbitmqConstPrepared, gin.H{
	// 	"gid": common.GenGid(),
	// })
	// common.PanicIfError(err)
	// queue := rb.QueueNew(dtmsvr.RabbitmqConstPrepared)
	// queue.WaitAndHandle(func(data map[string]interface{}) {
	// 	logrus.Printf("processed msg: %v in queue1", data)
	// })

	dtmsvr.Main()
	examples.Main()
	time.Sleep(1000 * 1000 * 1000 * 1000)
}

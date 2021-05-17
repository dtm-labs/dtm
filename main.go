package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmsvr"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	dtmsvr.LoadConfig()
	rb := dtmsvr.RabbitmqNew(&dtmsvr.ServerConfig.Rabbitmq)
	err := rb.SendAndConfirm(dtmsvr.RabbitmqConstPrepared, gin.H{
		"gid": common.GenGid(),
	})
	common.PanicIfError(err)
	queue := rb.QueueNew(dtmsvr.RabbitmqConstPrepared)
	queue.WaitAndHandle(func(data map[string]interface{}) {
		logrus.Printf("processed msg: %v in queue1", data)
	})
}

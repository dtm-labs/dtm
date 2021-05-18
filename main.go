package main

import (
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func main() {
	dtmsvr.LoadConfig()
	db := dtmsvr.DbGet()
	tx := db.Begin()
	common.PanicIfError(tx.Error)
	dbr := tx.Commit()
	common.PanicIfError(dbr.Error)

	tx = db.Begin()
	common.PanicIfError(tx.Error)
	dbr = tx.Commit()
	common.PanicIfError(dbr.Error)
	db.Exec("truncate test1.a_saga")
	db.Exec("truncate test1.a_saga_step")
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

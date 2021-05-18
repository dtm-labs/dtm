package main

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtm"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func init() {
	dtmsvr.LoadConfig()
}

func TestViper(t *testing.T) {
	assert.Equal(t, "test_val", viper.GetString("test"))
}

func TestDtmSvr(t *testing.T) {
	// 清理数据
	rabbit := dtmsvr.RabbitmqNew(&dtmsvr.ServerConfig.Rabbitmq)
	queprepared := rabbit.QueueNew(dtmsvr.RabbitmqConstPrepared)
	for i := 0; i < queprepared.Queue.Messages; i++ {
		queprepared.WaitAndHandleOne(func(data M) {
			logrus.Printf("ignoring prepared queue data before test")
		})
	}
	quecommited := rabbit.QueueNew(dtmsvr.RabbitmqConstCommited)
	for i := 0; i < quecommited.Queue.Messages; i++ {
		quecommited.WaitAndHandleOne(func(data M) {
			logrus.Printf("ignoring commited queue data before test")
		})
	}
	db := dtmsvr.DbGet()
	common.PanicIfError(db.Exec("truncate test1.a_saga").Error)
	common.PanicIfError(db.Exec("truncate test1.a_saga_step").Error)

	// 启动组件
	go dtmsvr.StartSvr()
	go examples.StartSvr()
	time.Sleep(time.Duration(100 * 1000 * 1000))

	// 开始第一个正常流程的测试
	saga := genSaga("gid-1", false, false)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	sm := dtmsvr.SagaModel{}
	db.Model(&sm).Where("gid=?", saga.Gid).First(&sm)
	assert.Equal(t, "prepared", sm.Status)
	saga.Commit()
	quecommited.WaitAndHandleOne(dtmsvr.HandleCommitedMsg)
	db.Model(&dtmsvr.SagaModel{}).Where("gid=?", saga.Gid).First(&sm)
	assert.Equal(t, "finished", sm.Status)
	steps := []dtmsvr.SagaStepModel{}
	db.Model(&dtmsvr.SagaStepModel{}).Where("gid=?", saga.Gid).Find(&steps)
	assert.Equal(t, true, steps[0].Status == "pending" && steps[2].Status == "pending" && steps[1].Status == "finished" && steps[3].Status == "finished")

	saga = genSaga("gid-2", false, true)
	saga.Commit()
	quecommited.WaitAndHandleOne(dtmsvr.HandleCommitedMsg)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	sm = dtmsvr.SagaModel{}
	db.Model(&dtmsvr.SagaModel{}).Where("gid=?", saga.Gid).First(&sm)
	assert.Equal(t, "rollbacked", sm.Status)
	steps = []dtmsvr.SagaStepModel{}
	db.Model(&dtmsvr.SagaStepModel{}).Where("gid=?", saga.Gid).Find(&steps)
	assert.Equal(t, true, steps[0].Status == "rollbacked" && steps[2].Status == "rollbacked" && steps[1].Status == "finished" && steps[3].Status == "rollbacked")

	// assert.Equal(t, 1, 0)
	// 开始测试

	// 发送Prepare请求后，验证数据库
	// ConsumeHalfMsg 验证数据库
	// ConsumeMsg 验证数据库
}

func genSaga(gid string, inFailed bool, outFailed bool) *dtm.Saga {
	saga := dtm.SagaNew(examples.TcServer, gid, examples.BusiApi+"/TransQuery")
	req := examples.TransReq{
		Amount:         30,
		TransInFailed:  inFailed,
		TransOutFailed: outFailed,
	}
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInCompensate", &req)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutCompensate", &req)
	return saga
}

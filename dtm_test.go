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

var myinit int = func() int {
	dtmsvr.LoadConfig()
	return 0
}()

// 测试使用的全局对象
var rabbit = dtmsvr.RabbitmqNew(&dtmsvr.ServerConfig.Rabbitmq)
var queprepared = rabbit.QueueNew(dtmsvr.RabbitmqConstPrepared)
var quecommited = rabbit.QueueNew(dtmsvr.RabbitmqConstCommited)
var db = dtmsvr.DbGet()

func getSagaModel(gid string) *dtmsvr.SagaModel {
	sm := dtmsvr.SagaModel{}
	dbr := db.Model(&sm).Where("gid=?", gid).First(&sm)
	common.PanicIfError(dbr.Error)
	return &sm
}

func getSagaStepStatus(gid string) []string {
	steps := []dtmsvr.SagaStepModel{}
	dbr := db.Model(&dtmsvr.SagaStepModel{}).Where("gid=?", gid).Find(&steps)
	common.PanicIfError(dbr.Error)
	status := []string{}
	for _, step := range steps {
		status = append(status, step.Status)
	}
	return status
}

func noramlSaga(t *testing.T) {
	saga := genSaga("gid-normal", false, false)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	saga.Commit()
	quecommited.WaitAndHandleOne(dtmsvr.HandleCommitedMsg)
	assert.Equal(t, "finished", getSagaModel(saga.Gid).Status)
	assert.Equal(t, []string{"pending", "finished", "pending", "finished"}, getSagaStepStatus(saga.Gid))
}

func rollbackSaga2(t *testing.T) {
	saga := genSaga("gid-rollback2", false, true)
	saga.Commit()
	quecommited.WaitAndHandleOne(dtmsvr.HandleCommitedMsg)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	assert.Equal(t, "rollbacked", getSagaModel(saga.Gid).Status)
	assert.Equal(t, []string{"rollbacked", "finished", "rollbacked", "rollbacked"}, getSagaStepStatus(saga.Gid))
}

func prepareCancel(t *testing.T) {
	saga := genSaga("gid1-trans-cancel", false, true)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	dtmsvr.CronPreparedOne(-1 * time.Second)
	assert.Equal(t, "canceled", getSagaModel(saga.Gid).Status)
}

func preparePending(t *testing.T) {
	saga := genSaga("gid1-trans-pending", false, true)
	saga.Prepare()
	queprepared.WaitAndHandleOne(dtmsvr.HandlePreparedMsg)
	dtmsvr.CronPreparedOne(-1 * time.Second)
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
}

func TestDtmSvr(t *testing.T) {
	// 清理数据
	for i := 0; i < queprepared.Queue.Messages; i++ {
		queprepared.WaitAndHandleOne(func(data M) {
			logrus.Printf("ignoring prepared queue data before test")
		})
	}
	for i := 0; i < quecommited.Queue.Messages; i++ {
		quecommited.WaitAndHandleOne(func(data M) {
			logrus.Printf("ignoring commited queue data before test")
		})
	}
	common.PanicIfError(db.Exec("truncate test1.a_saga").Error)
	common.PanicIfError(db.Exec("truncate test1.a_saga_step").Error)

	// 启动组件
	go dtmsvr.StartSvr()
	go examples.StartSvr()
	time.Sleep(time.Duration(100 * 1000 * 1000))

	prepareCancel(t)
	preparePending(t)
	noramlSaga(t)
	rollbackSaga2(t)
	// assert.Equal(t, 1, 0)
	// 开始测试

	// 发送Prepare请求后，验证数据库
	// ConsumeHalfMsg 验证数据库
	// ConsumeMsg 验证数据库
}

func genSaga(gid string, inFailed bool, outFailed bool) *dtm.Saga {
	saga := dtm.SagaNew(examples.TcServer, gid, examples.Busi+"/TransQuery")
	req := examples.TransReq{
		Amount:         30,
		TransInFailed:  inFailed,
		TransOutFailed: outFailed,
	}
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInCompensate", &req)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutCompensate", &req)
	return saga
}

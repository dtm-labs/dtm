package dtmsvr

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yedf/dtm"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/examples"
)

var myinit int = func() int {
	LoadConfig()
	return 0
}()

func TestViper(t *testing.T) {
	assert.Equal(t, true, viper.Get("mysql") != nil)
	assert.Equal(t, int64(90), Config.PreparedExpire)
}

func TestDtmSvr(t *testing.T) {
	SagaProcessedTestChan = make(chan string, 1)
	// 清理数据
	common.PanicIfError(db.Exec("truncate saga").Error)
	common.PanicIfError(db.Exec("truncate saga_step").Error)
	common.PanicIfError(db.Exec("truncate trans_log").Error)

	// 启动组件
	go StartSvr()
	go examples.StartSvr()
	time.Sleep(time.Duration(100 * 1000 * 1000))

	preparePending(t)
	prepareCancel(t)
	commitedPending(t)
	noramlSaga(t)
	rollbackSaga2(t)
}

func TestCover(t *testing.T) {
	db := DbGet()
	db.NoMust()
	CronPreparedOnce(0)
	CronCommitedOnce(0)
	defer handlePanic()
	checkAffected(db.DB)
}

// 测试使用的全局对象
var initdb = DbGet()

func getSagaModel(gid string) *SagaModel {
	sm := SagaModel{}
	dbr := db.Model(&sm).Where("gid=?", gid).First(&sm)
	common.PanicIfError(dbr.Error)
	return &sm
}

func getSagaStepStatus(gid string) []string {
	steps := []SagaStepModel{}
	dbr := db.Model(&SagaStepModel{}).Where("gid=?", gid).Find(&steps)
	common.PanicIfError(dbr.Error)
	status := []string{}
	for _, step := range steps {
		status = append(status, step.Status)
	}
	return status
}

func noramlSaga(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Prepare()
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	saga.Commit()
	assert.Equal(t, "commited", getSagaModel(saga.Gid).Status)
	WaitCommitedSaga(saga.Gid)
	assert.Equal(t, []string{"pending", "finished", "pending", "finished"}, getSagaStepStatus(saga.Gid))
}

func rollbackSaga2(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Commit()
	WaitCommitedSaga(saga.Gid)
	saga.Prepare()
	assert.Equal(t, "rollbacked", getSagaModel(saga.Gid).Status)
	assert.Equal(t, []string{"rollbacked", "finished", "rollbacked", "rollbacked"}, getSagaStepStatus(saga.Gid))
}

func prepareCancel(t *testing.T) {
	saga := genSaga("gid1-prepareCancel", false, true)
	saga.Prepare()
	examples.TransQueryResult = "FAIL"
	Config.PreparedExpire = 0
	CronPreparedOnce(-10 * time.Second)
	examples.TransQueryResult = ""
	Config.PreparedExpire = 60
	assert.Equal(t, "canceled", getSagaModel(saga.Gid).Status)
}

func preparePending(t *testing.T) {
	saga := genSaga("gid1-preparePending", false, false)
	saga.Prepare()
	examples.TransQueryResult = "PENDING"
	CronPreparedOnce(-10 * time.Second)
	examples.TransQueryResult = ""
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	CronPreparedOnce(-10 * time.Second)
	WaitCommitedSaga(saga.Gid)
	assert.Equal(t, "finished", getSagaModel(saga.Gid).Status)
}

func commitedPending(t *testing.T) {
	saga := genSaga("gid-commitedPending", false, false)
	saga.Prepare()
	saga.Commit()
	examples.TransOutResult = "PENDING"
	WaitCommitedSaga(saga.Gid)
	assert.Equal(t, []string{"pending", "finished", "pending", "pending"}, getSagaStepStatus(saga.Gid))
	examples.TransOutResult = ""
	CronCommitedOnce(-10 * time.Second)
	WaitCommitedSaga(saga.Gid)
	assert.Equal(t, []string{"pending", "finished", "pending", "finished"}, getSagaStepStatus(saga.Gid))
	assert.Equal(t, "finished", getSagaModel(saga.Gid).Status)
}

func genSaga(gid string, inFailed bool, outFailed bool) *dtm.Saga {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	saga := dtm.SagaNew(examples.DtmServer, gid, examples.Busi+"/TransQuery")
	req := examples.TransReq{
		Amount:         30,
		TransInFailed:  inFailed,
		TransOutFailed: outFailed,
	}
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInCompensate", &req)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutCompensate", &req)
	return saga
}

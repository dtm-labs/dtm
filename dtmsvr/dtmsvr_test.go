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
	common.InitApp(&config)
	return 0
}()

func TestViper(t *testing.T) {
	assert.Equal(t, true, viper.Get("mysql") != nil)
	assert.Equal(t, int64(90), config.PreparedExpire)
}

func TestDtmSvr(t *testing.T) {
	TransProcessedTestChan = make(chan string, 1)

	// 启动组件
	go StartSvr()
	go examples.SagaStartSvr()
	go examples.XaStartSvr()
	time.Sleep(time.Duration(100 * 1000 * 1000))

	// 清理数据
	common.PanicIfError(dbGet().Exec("truncate trans_global").Error)
	common.PanicIfError(dbGet().Exec("truncate trans_branch").Error)
	common.PanicIfError(dbGet().Exec("truncate trans_log").Error)
	examples.ResetXaData()

	xaRollback(t)
	xaNormal(t)
	sagaPreparePending(t)
	sagaPrepareCancel(t)
	sagaCommittedPending(t)
	sagaNormal(t)
	sagaRollback(t)
}

func TestCover(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronPreparedOnce(0)
	CronCommittedOnce(0)
	defer handlePanic()
	checkAffected(db.DB)
}

// 测试使用的全局对象
var initdb = dbGet()

func getSagaModel(gid string) *TransGlobalModel {
	sm := TransGlobalModel{}
	dbr := dbGet().Model(&sm).Where("gid=?", gid).First(&sm)
	common.PanicIfError(dbr.Error)
	return &sm
}

func getBranchesStatus(gid string) []string {
	steps := []TransBranchModel{}
	dbr := dbGet().Model(&TransBranchModel{}).Where("gid=?", gid).Find(&steps)
	common.PanicIfError(dbr.Error)
	status := []string{}
	for _, step := range steps {
		status = append(status, step.Status)
	}
	return status
}

func xaNormal(t *testing.T) {
	xa := examples.XaClient
	gid := "xa-normal"
	err := xa.XaGlobalTransaction(gid, func() error {
		req := examples.GenTransReq(30, false, false)
		resp, err := common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "1",
		}).Post(examples.XaBusi + "/TransOut")
		common.CheckRestySuccess(resp, err)
		resp, err = common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "2",
		}).Post(examples.XaBusi + "/TransIn")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	common.PanicIfError(err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"finished", "finished"}, getBranchesStatus(gid))
}

func xaRollback(t *testing.T) {
	xa := examples.XaClient
	gid := "xa-rollback"
	err := xa.XaGlobalTransaction(gid, func() error {
		req := examples.GenTransReq(30, false, true)
		resp, err := common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "1",
		}).Post(examples.XaBusi + "/TransOut")
		common.CheckRestySuccess(resp, err)
		resp, err = common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "2",
		}).Post(examples.XaBusi + "/TransIn")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	if err != nil {
		logrus.Errorf("global transaction failed, so rollback")
	}
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"rollbacked"}, getBranchesStatus(gid))
}

func sagaNormal(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Prepare()
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	saga.Commit()
	assert.Equal(t, "committed", getSagaModel(saga.Gid).Status)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "finished", "prepared", "finished"}, getBranchesStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	saga.Prepare()
	assert.Equal(t, "rollbacked", getSagaModel(saga.Gid).Status)
	assert.Equal(t, []string{"rollbacked", "finished", "rollbacked", "rollbacked"}, getBranchesStatus(saga.Gid))
}

func sagaPrepareCancel(t *testing.T) {
	saga := genSaga("gid1-prepareCancel", false, true)
	saga.Prepare()
	examples.TransQueryResult = "FAIL"
	config.PreparedExpire = -10
	CronPreparedOnce(-10 * time.Second)
	examples.TransQueryResult = ""
	config.PreparedExpire = 60
	assert.Equal(t, "canceled", getSagaModel(saga.Gid).Status)
}

func sagaPreparePending(t *testing.T) {
	saga := genSaga("gid1-preparePending", false, false)
	saga.Prepare()
	examples.TransQueryResult = "PENDING"
	CronPreparedOnce(-10 * time.Second)
	examples.TransQueryResult = ""
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	CronPreparedOnce(-10 * time.Second)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "finished", getSagaModel(saga.Gid).Status)
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	saga.Prepare()
	examples.TransInResult = "PENDING"
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	examples.TransInResult = ""
	assert.Equal(t, []string{"prepared", "finished", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronCommittedOnce(-10 * time.Second)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "finished", "prepared", "finished"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "finished", getSagaModel(saga.Gid).Status)
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtm.Saga {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	saga := dtm.SagaNew(examples.DtmServer, gid, examples.SagaBusi+"/TransQuery")
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.SagaBusi+"/TransOut", examples.SagaBusi+"/TransOutCompensate", &req)
	saga.Add(examples.SagaBusi+"/TransIn", examples.SagaBusi+"/TransInCompensate", &req)
	return saga
}

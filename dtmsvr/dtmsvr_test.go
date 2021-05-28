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
	e2p(dbGet().Exec("truncate trans_global").Error)
	e2p(dbGet().Exec("truncate trans_branch").Error)
	e2p(dbGet().Exec("truncate trans_log").Error)
	examples.ResetXaData()
	// tccNormal(t)
	sagaCommittedPending(t)
	sagaPreparePending(t)
	xaRollback(t)
	xaNormal(t)
	sagaPrepareCancel(t)
	sagaNormal(t)
	sagaRollback(t)
}

func TestCover(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronTransOnce(0, "prepared")
	CronTransOnce(0, "committed")
	defer handlePanic()
	checkAffected(db.DB)
}

// 测试使用的全局对象
var initdb = dbGet()

func getSagaModel(gid string) *TransGlobal {
	sm := TransGlobal{}
	dbr := dbGet().Model(&sm).Where("gid=?", gid).First(&sm)
	e2p(dbr.Error)
	return &sm
}

func getBranchesStatus(gid string) []string {
	steps := []TransBranch{}
	dbr := dbGet().Model(&TransBranch{}).Where("gid=?", gid).Find(&steps)
	e2p(dbr.Error)
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
	e2p(err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(gid))
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
	assert.Equal(t, []string{"failed"}, getBranchesStatus(gid))
}

func tccNormal(t *testing.T) {
	tcc := genTcc("gid-normal-tcc", false, false)
	tcc.Prepare(tcc.QueryPrepared)
	assert.Equal(t, "prepared", getSagaModel(tcc.Gid).Status)
	tcc.Commit()
	assert.Equal(t, "committed", getSagaModel(tcc.Gid).Status)
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "succeed", "prepared", "succeed", "succeed"}, getBranchesStatus(tcc.Gid))

}
func sagaNormal(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Prepare(saga.QueryPrepared)
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	saga.Commit()
	assert.Equal(t, "committed", getSagaModel(saga.Gid).Status)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	saga.Prepare(saga.QueryPrepared)
	assert.Equal(t, "failed", getSagaModel(saga.Gid).Status)
	assert.Equal(t, []string{"failed", "succeed", "failed", "failed"}, getBranchesStatus(saga.Gid))
}

func sagaPrepareCancel(t *testing.T) {
	saga := genSaga("gid1-prepareCancel", false, true)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransQueryResult = "FAIL"
	config.PreparedExpire = -10
	CronTransOnce(-10*time.Second, "prepared")
	examples.SagaTransQueryResult = ""
	config.PreparedExpire = 60
	assert.Equal(t, "canceled", getSagaModel(saga.Gid).Status)
}

func sagaPreparePending(t *testing.T) {
	saga := genSaga("gid1-preparePending", false, false)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransQueryResult = "PENDING"
	CronTransOnce(-10*time.Second, "prepared")
	examples.SagaTransQueryResult = ""
	assert.Equal(t, "prepared", getSagaModel(saga.Gid).Status)
	CronTransOnce(-10*time.Second, "prepared")
	assert.Equal(t, "succeed", getSagaModel(saga.Gid).Status)
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransInResult = "PENDING"
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	examples.SagaTransInResult = ""
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(-10*time.Second, "committed")
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getSagaModel(saga.Gid).Status)
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtm.Saga {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	saga := dtm.SagaNew(examples.DtmServer, gid)
	saga.QueryPrepared = examples.SagaBusi + "/TransQuery"
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.SagaBusi+"/TransOut", examples.SagaBusi+"/TransOutCompensate", &req)
	saga.Add(examples.SagaBusi+"/TransIn", examples.SagaBusi+"/TransInCompensate", &req)
	return saga
}

func genTcc(gid string, outFailed bool, inFailed bool) *dtm.Tcc {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	tcc := dtm.TccNew(examples.DtmServer, gid)
	tcc.QueryPrepared = examples.TccBusi + "/TransQuery"
	req := examples.GenTransReq(30, outFailed, inFailed)
	tcc.Add(examples.TccBusi+"/TransOutTry", examples.TccBusi+"/TransOutConfirm", examples.TccBusi+"/TransOutCancel", &req)
	tcc.Add(examples.TccBusi+"/TransInTry", examples.TccBusi+"/TransInConfirm", examples.TccBusi+"/TransInCancel", &req)
	return tcc
}

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
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
	return 0
}()

func TestViper(t *testing.T) {
	assert.Equal(t, true, viper.Get("mysql") != nil)
	assert.Equal(t, int64(90), config.PreparedExpire)
}

func TestDtmSvr(t *testing.T) {
	TransProcessedTestChan = make(chan string, 1)
	PopulateMysql()
	examples.PopulateMysql()
	// 启动组件
	go StartSvr()
	go examples.SagaStartSvr()
	go examples.XaStartSvr()
	go examples.TccStartSvr()
	go examples.MsgStartSvr()
	time.Sleep(time.Duration(200 * 1000 * 1000))

	// 清理数据
	e2p(dbGet().Exec("truncate trans_global").Error)
	e2p(dbGet().Exec("truncate trans_branch").Error)
	e2p(dbGet().Exec("truncate trans_log").Error)
	examples.ResetXaData()

	msgPending(t)
	msgNormal(t)
	sagaNormal(t)
	tccNormal(t)
	tccRollback(t)
	tccRollbackPending(t)
	xaNormal(t)
	xaRollback(t)
	sagaCommittedPending(t)
	sagaPreparePending(t)
	sagaPrepareCancel(t)
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

func getTransStatus(gid string) string {
	sm := TransGlobal{}
	dbr := dbGet().Model(&sm).Where("gid=?", gid).First(&sm)
	e2p(dbr.Error)
	return sm.Status
}

func getBranchesStatus(gid string) []string {
	branches := []TransBranch{}
	dbr := dbGet().Model(&TransBranch{}).Where("gid=?", gid).Order("id").Find(&branches)
	e2p(dbr.Error)
	status := []string{}
	for _, branch := range branches {
		status = append(status, branch.Status)
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
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(gid))
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
	assert.Equal(t, []string{"succeed", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}

func tccNormal(t *testing.T) {
	tcc := genTcc("gid-tcc-normal", false, false)
	tcc.Prepare(tcc.QueryPrepared)
	assert.Equal(t, "prepared", getTransStatus(tcc.Gid))
	tcc.Commit()
	assert.Equal(t, "committed", getTransStatus(tcc.Gid))
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "succeed", "prepared", "succeed", "succeed"}, getBranchesStatus(tcc.Gid))
}
func tccRollback(t *testing.T) {
	tcc := genTcc("gid-tcc-rollback", false, true)
	tcc.Commit()
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, []string{"succeed", "prepared", "succeed", "succeed", "prepared", "failed"}, getBranchesStatus(tcc.Gid))
}
func tccRollbackPending(t *testing.T) {
	tcc := genTcc("gid-tcc-rollback-pending", false, true)
	examples.TccTransInCancelResult = "PENDING"
	tcc.Commit()
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, "committed", getTransStatus(tcc.Gid))
	examples.TccTransInCancelResult = ""
	CronTransOnce(60*time.Second, "committed")
	assert.Equal(t, []string{"succeed", "prepared", "succeed", "succeed", "prepared", "failed"}, getBranchesStatus(tcc.Gid))
}

func msgNormal(t *testing.T) {
	msg := genMsg("gid-normal-msg")
	msg.Commit()
	assert.Equal(t, "committed", getTransStatus(msg.Gid))
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(msg.Gid))
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func msgPending(t *testing.T) {
	msg := genMsg("gid-normal-pending")
	msg.Prepare("")
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MsgTransQueryResult = "PENDING"
	CronTransOnce(60*time.Second, "prepared")
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MsgTransQueryResult = ""
	examples.MsgTransInResult = "PENDING"
	CronTransOnce(60*time.Second, "prepared")
	assert.Equal(t, "committed", getTransStatus(msg.Gid))
	examples.MsgTransInResult = ""
	CronTransOnce(60*time.Second, "committed")
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(msg.Gid))
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func sagaNormal(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Prepare(saga.QueryPrepared)
	assert.Equal(t, "prepared", getTransStatus(saga.Gid))
	saga.Commit()
	assert.Equal(t, "committed", getTransStatus(saga.Gid))
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	saga.Prepare(saga.QueryPrepared)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

func sagaPrepareCancel(t *testing.T) {
	saga := genSaga("gid1-prepareCancel", false, true)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransQueryResult = "FAIL"
	config.PreparedExpire = -10
	CronTransOnce(60*time.Second, "prepared")
	examples.SagaTransQueryResult = ""
	config.PreparedExpire = 60
	assert.Equal(t, "canceled", getTransStatus(saga.Gid))
}

func sagaPreparePending(t *testing.T) {
	saga := genSaga("gid1-preparePending", false, false)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransQueryResult = "PENDING"
	CronTransOnce(60*time.Second, "prepared")
	examples.SagaTransQueryResult = ""
	assert.Equal(t, "prepared", getTransStatus(saga.Gid))
	CronTransOnce(60*time.Second, "prepared")
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	saga.Prepare(saga.QueryPrepared)
	examples.SagaTransInResult = "PENDING"
	saga.Commit()
	WaitTransProcessed(saga.Gid)
	examples.SagaTransInResult = ""
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(60*time.Second, "committed")
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func genMsg(gid string) *dtm.Msg {
	logrus.Printf("beginning a msg test ---------------- %s", gid)
	msg := dtm.MsgNew(examples.DtmServer, gid)
	msg.QueryPrepared = examples.MsgBusi + "/TransQuery"
	req := examples.GenTransReq(30, false, false)
	msg.Add(examples.MsgBusi+"/TransOut", &req)
	msg.Add(examples.MsgBusi+"/TransIn", &req)
	return msg
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
	logrus.Printf("beginning a tcc test ---------------- %s", gid)
	tcc := dtm.TccNew(examples.DtmServer, gid)
	tcc.QueryPrepared = examples.TccBusi + "/TransQuery"
	req := examples.GenTransReq(30, outFailed, inFailed)
	tcc.Add(examples.TccBusi+"/TransOutTry", examples.TccBusi+"/TransOutConfirm", examples.TccBusi+"/TransOutCancel", &req)
	tcc.Add(examples.TccBusi+"/TransInTry", examples.TccBusi+"/TransInConfirm", examples.TccBusi+"/TransInCancel", &req)
	return tcc
}

func transQuery(t *testing.T, gid string) {
	resp, err := common.RestyClient.R().SetQueryParam("gid", gid).Get(examples.DtmServer + "/query")
	e2p(err)
	m := M{}
	assert.Equal(t, resp.StatusCode(), 200)
	common.MustUnmarshalString(resp.String(), &m)
	assert.NotEqual(t, nil, m["transaction"])
	assert.Equal(t, 4, len(m["branches"].([]interface{})))

	resp, err = common.RestyClient.R().SetQueryParam("gid", "").Get(examples.DtmServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 500)

	resp, err = common.RestyClient.R().SetQueryParam("gid", "1").Get(examples.DtmServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 200)
	common.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, nil, m["transaction"])
	assert.Equal(t, 0, len(m["branches"].([]interface{})))
}

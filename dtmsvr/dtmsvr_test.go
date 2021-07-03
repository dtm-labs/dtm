package dtmsvr

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

var myinit int = func() int {
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
	PopulateMysql()
	examples.PopulateMysql()
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
	app := examples.BaseAppNew()
	examples.BaseAppSetup(app)
	examples.SagaSetup(app)
	examples.TccSetup(app)
	examples.XaSetup(app)
	examples.MsgSetup(app)
	examples.BaseAppStart(app)

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
	sagaRollback(t)

}

func TestCover(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronTransOnce(0, "prepared")
	CronTransOnce(0, "submitted")
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
	gid, err := xa.XaGlobalTransaction(func(gid string) error {
		req := examples.GenTransReq(30, false, false)
		resp, err := common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "1",
		}).Post(examples.Busi + "/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "2",
		}).Post(examples.Busi + "/TransInXa")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	e2p(err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(gid))
}

func xaRollback(t *testing.T) {
	xa := examples.XaClient
	gid, err := xa.XaGlobalTransaction(func(gid string) error {
		req := examples.GenTransReq(30, false, true)
		resp, err := common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "1",
		}).Post(examples.Busi + "/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "2",
		}).Post(examples.Busi + "/TransInXa")
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
	tcc.Submit()
	assert.Equal(t, "submitted", getTransStatus(tcc.Gid))
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "succeed", "prepared", "succeed", "succeed"}, getBranchesStatus(tcc.Gid))
}
func tccRollback(t *testing.T) {
	tcc := genTcc("gid-tcc-rollback", false, true)
	tcc.Submit()
	WaitTransProcessed(tcc.Gid)
	assert.Equal(t, []string{"succeed", "prepared", "succeed", "succeed", "prepared", "failed"}, getBranchesStatus(tcc.Gid))
}
func tccRollbackPending(t *testing.T) {
	tcc := genTcc("gid-tcc-rollback-pending", false, true)
	examples.MainSwitch.TransInRevertResult.SetOnce("PENDING")
	tcc.Submit()
	WaitTransProcessed(tcc.Gid)
	// assert.Equal(t, "submitted", getTransStatus(tcc.Gid))
	CronTransOnce(60*time.Second, "submitted")
	assert.Equal(t, []string{"succeed", "prepared", "succeed", "succeed", "prepared", "failed"}, getBranchesStatus(tcc.Gid))
}

func msgNormal(t *testing.T) {
	msg := genMsg("gid-normal-msg")
	msg.Submit()
	assert.Equal(t, "submitted", getTransStatus(msg.Gid))
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(msg.Gid))
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func msgPending(t *testing.T) {
	msg := genMsg("gid-normal-pending")
	msg.Prepare("")
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce("PENDING")
	CronTransOnce(60*time.Second, "prepared")
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	CronTransOnce(60*time.Second, "prepared")
	assert.Equal(t, "submitted", getTransStatus(msg.Gid))
	CronTransOnce(60*time.Second, "submitted")
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(msg.Gid))
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func sagaNormal(t *testing.T) {
	saga := genSaga("gid-noramlSaga", false, false)
	saga.Submit()
	assert.Equal(t, "submitted", getTransStatus(saga.Gid))
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	transQuery(t, saga.Gid)
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "prepared", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(60*time.Second, "submitted")
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func genMsg(gid string) *dtmcli.Msg {
	logrus.Printf("beginning a msg test ---------------- %s", gid)
	msg := dtmcli.NewMsg(examples.DtmServer)
	msg.QueryPrepared = examples.Busi + "/CanSubmit"
	req := examples.GenTransReq(30, false, false)
	msg.Add(examples.Busi+"/TransOut", &req)
	msg.Add(examples.Busi+"/TransIn", &req)
	msg.Gid = gid
	return msg
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	saga := dtmcli.NewSaga(examples.DtmServer)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
	saga.Gid = gid
	return saga
}

func genTcc(gid string, outFailed bool, inFailed bool) *dtmcli.Tcc {
	logrus.Printf("beginning a tcc test ---------------- %s", gid)
	tcc := dtmcli.NewTcc(examples.DtmServer)
	req := examples.GenTransReq(30, outFailed, inFailed)
	tcc.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutConfirm", examples.Busi+"/TransOutRevert", &req)
	tcc.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInConfirm", examples.Busi+"/TransInRevert", &req)
	tcc.Gid = gid
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

func TestSqlDB(t *testing.T) {
	asserts := assert.New(t)
	db := common.DbGet(config.Mysql)
	db.Must().Exec("insert ignore into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type) values('saga', 'gid1', 'branch_id1', 'action')")
	_, err := dtmcli.ThroughBarrierCall(db.ToSqlDB(), "saga", "gid2", "branch_id2", "compensate", func(db *sql.DB) (interface{}, error) {
		logrus.Printf("rollback gid2")
		return nil, fmt.Errorf("gid2 error")
	})
	asserts.Error(err, fmt.Errorf("gid2 error"))
	dbr := db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid1").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(0))
	_, err = dtmcli.ThroughBarrierCall(db.ToSqlDB(), "saga", "gid2", "branch_id2", "compensate", func(db *sql.DB) (interface{}, error) {
		logrus.Printf("submit gid2")
		return nil, nil
	})
	asserts.Nil(err)
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(2))
}

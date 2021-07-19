package dtmsvr

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

var DtmServer = examples.DtmServer
var Busi = examples.Busi
var app *gin.Engine

func init() {
	TransProcessedTestChan = make(chan string, 1)
	common.InitApp(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
	PopulateMysql()
	examples.PopulateMysql()
	// 启动组件
	go StartSvr()
	app = examples.BaseAppStartup()
	examples.SagaSetup(app)
	examples.TccSetup(app)
	examples.XaSetup(app)
	examples.MsgSetup(app)
	examples.TccBarrierAddRoute(app)
	examples.SagaBarrierAddRoute(app)

	// 清理数据
	e2p(dbGet().Exec("truncate trans_global").Error)
	e2p(dbGet().Exec("truncate trans_branch").Error)
	e2p(dbGet().Exec("truncate trans_log").Error)
	examples.ResetXaData()
}

func TestDtmSvr(t *testing.T) {

	tccBarrierDisorder(t)
	tccBarrierNormal(t)
	tccBarrierRollback(t)
	sagaBarrierNormal(t)
	sagaBarrierRollback(t)
	msgNormal(t)
	msgPending(t)
	tccNormal(t)
	tccRollback(t)
	sagaNormal(t)
	xaNormal(t)
	xaRollback(t)
	sagaCommittedPending(t)
	sagaRollback(t)

	// for coverage
	examples.QsStartSvr()
	assertSucceed(t, examples.QsFireRequest())
	assertSucceed(t, examples.MsgFireRequest())
	assertSucceed(t, examples.SagaBarrierFireRequest())
	assertSucceed(t, examples.SagaFireRequest())
	assertSucceed(t, examples.TccBarrierFireRequest())
	assertSucceed(t, examples.TccFireRequest())
	assertSucceed(t, examples.XaFireRequest())
}

func TestCover(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronTransOnce(0)
	defer handlePanic()
	checkAffected(db.DB)

	go CronExpiredTrans(1)
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
	xc := examples.XaClient
	gid, err := xc.XaGlobalTransaction(func(xa *dtmcli.Xa) error {
		req := examples.GenTransReq(30, false, false)
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = xa.CallBranch(req, examples.Busi+"/TransInXa")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	e2p(err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(gid))
}

func xaRollback(t *testing.T) {
	xc := examples.XaClient
	gid, err := xc.XaGlobalTransaction(func(xa *dtmcli.Xa) error {
		req := examples.GenTransReq(30, false, true)
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		common.CheckRestySuccess(resp, err)
		resp, err = xa.CallBranch(req, examples.Busi+"/TransInXa")
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
	data := &examples.TransReq{Amount: 30}
	_, err := dtmcli.TccGlobalTransaction(examples.DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		_, rerr = tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return
	})
	e2p(err)
}
func tccBarrierNormal(t *testing.T) {
	_, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		e2p(rerr)
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		e2p(rerr)
		if res2.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res2.StatusCode())
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	e2p(err)
}

func tccBarrierRollback(t *testing.T) {
	gid, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		e2p(rerr)
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		e2p(rerr)
		if res2.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res2.StatusCode())
		}
		if strings.Contains(res2.String(), "FAILURE") {
			return fmt.Errorf("branch trans in fail")
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	assert.Equal(t, err, fmt.Errorf("branch trans in fail"))
	WaitTransProcessed(gid)
	assert.Equal(t, "failed", getTransStatus(gid))
}

func tccRollback(t *testing.T) {
	data := &examples.TransReq{Amount: 30, TransInResult: "FAILURE"}
	_, err := dtmcli.TccGlobalTransaction(examples.DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		_, rerr = tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return
	})
	e2p(err)
}
func msgNormal(t *testing.T) {
	msg := genMsg("gid-msg-normal")
	msg.Submit()
	assert.Equal(t, "submitted", getTransStatus(msg.Gid))
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, []string{"succeed", "succeed"}, getBranchesStatus(msg.Gid))
	assert.Equal(t, "succeed", getTransStatus(msg.Gid))
}

func msgPending(t *testing.T) {
	msg := genMsg("gid-msg-normal-pending")
	msg.Prepare("")
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "prepared", getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "submitted", getTransStatus(msg.Gid))
	CronTransOnce(60 * time.Second)
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

func sagaBarrierNormal(t *testing.T) {
	req := &examples.TransReq{Amount: 30}
	saga := dtmcli.NewSaga(DtmServer).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
	logrus.Printf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
}

func sagaRollback(t *testing.T) {
	saga := genSaga("gid-rollbackSaga2", false, true)
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
	assert.Equal(t, []string{"succeed", "succeed", "succeed", "failed"}, getBranchesStatus(saga.Gid))
}

func sagaBarrierRollback(t *testing.T) {
	saga := dtmcli.NewSaga(DtmServer).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", &examples.TransReq{Amount: 30}).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", &examples.TransReq{Amount: 30, TransInResult: "FAILURE"})
	logrus.Printf("busi trans submit")
	err := saga.Submit()
	e2p(err)
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, "failed", getTransStatus(saga.Gid))
}

func sagaCommittedPending(t *testing.T) {
	saga := genSaga("gid-committedPending", false, false)
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	saga.Submit()
	WaitTransProcessed(saga.Gid)
	assert.Equal(t, []string{"prepared", "prepared", "prepared", "prepared"}, getBranchesStatus(saga.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(saga.Gid))
	assert.Equal(t, "succeed", getTransStatus(saga.Gid))
}

func assertSucceed(t *testing.T, gid string) {
	WaitTransProcessed(gid)
	assert.Equal(t, "succeed", getTransStatus(gid))
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
	transInfo := &dtmcli.TransInfo{
		TransType:  "saga",
		Gid:        "gid2",
		BranchID:   "branch_id2",
		BranchType: "action",
	}
	db.Must().Exec("insert ignore into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type, reason) values('saga', 'gid1', 'branch_id1', 'action', 'saga')")
	_, err := dtmcli.ThroughBarrierCall(db.ToSQLDB(), transInfo, func(db *sql.DB) (interface{}, error) {
		logrus.Printf("rollback gid2")
		return nil, fmt.Errorf("gid2 error")
	})
	asserts.Error(err, fmt.Errorf("gid2 error"))
	dbr := db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid1").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(0))
	_, err = dtmcli.ThroughBarrierCall(db.ToSQLDB(), transInfo, func(db *sql.DB) (interface{}, error) {
		logrus.Printf("submit gid2")
		return nil, nil
	})
	asserts.Nil(err)
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
}

func tccBarrierDisorder(t *testing.T) {
	timeoutChan := make(chan string, 2)
	finishedChan := make(chan string, 2)
	gid, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		body := &examples.TransReq{Amount: 30}
		tryURL := Busi + "/TccBTransOutTry"
		confirmURL := Busi + "/TccBTransOutConfirm"
		cancelURL := Busi + "/TccBSleepCancel"
		// 请参见子事务屏障里的时序图，这里为了模拟该时序图，手动拆解了callbranch
		branchID := tcc.NewBranchID()
		sleeped := false
		app.POST(examples.BusiAPI+"/TccBSleepCancel", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			res, err := examples.TccBarrierTransOutCancel(c)
			if !sleeped {
				sleeped = true
				logrus.Printf("sleep before cancel return")
				<-timeoutChan
				finishedChan <- "1"
			}
			return res, err
		}))
		// 注册子事务
		_, err := common.RestyClient.R().
			SetBody(&M{
				"gid":        tcc.Gid,
				"branch_id":  branchID,
				"trans_type": "tcc",
				"status":     "prepared",
				"data":       string(common.MustMarshal(body)),
				"try":        tryURL,
				"confirm":    confirmURL,
				"cancel":     cancelURL,
			}).
			Post(tcc.Dtm + "/registerTccBranch")
		e2p(err)
		go func() {
			logrus.Printf("sleeping to wait for tcc try timeout")
			<-timeoutChan
			_, _ = common.RestyClient.R().
				SetBody(body).
				SetQueryParams(common.MS{
					"dtm":         tcc.Dtm,
					"gid":         tcc.Gid,
					"branch_id":   branchID,
					"trans_type":  "tcc",
					"branch_type": "try",
				}).
				Post(tryURL)
			finishedChan <- "1"
		}()
		logrus.Printf("cron to timeout and then call cancel")
		go CronTransOnce(60 * time.Second)
		time.Sleep(100 * time.Millisecond)
		logrus.Printf("cron to timeout and then call cancelled twice")
		CronTransOnce(60 * time.Second)
		timeoutChan <- "wake"
		timeoutChan <- "wake"
		<-finishedChan
		<-finishedChan
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{"succeed", "prepared", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}

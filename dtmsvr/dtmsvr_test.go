package dtmsvr

import (
	"database/sql"
	"fmt"
	"testing"

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

func TestMain(m *testing.M) {
	TransProcessedTestChan = make(chan string, 1)
	common.InitConfig(common.GetProjectDir(), &config)
	PopulateDB(false)
	examples.PopulateDB(false)
	// 启动组件
	go StartSvr()
	app = examples.BaseAppStartup()
	examples.SagaSetup(app)
	examples.TccSetup(app)
	examples.XaSetup(app)
	examples.MsgSetup(app)
	examples.TccBarrierAddRoute(app)
	examples.SagaBarrierAddRoute(app)

	examples.ResetXaData()
	m.Run()
}

func TestCover(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronTransOnce(0)
	err := common.CatchP(func() {
		checkAffected(db.DB)
	})
	assert.Error(t, err)

	CronExpiredTrans(1)
	go sleepCronTime()
}

func TestType(t *testing.T) {
	err := common.CatchP(func() {
		dtmcli.MustGenGid("http://localhost:8080/api/no")
	})
	assert.Error(t, err)
	err = common.CatchP(func() {
		resp, err := common.RestyClient.R().SetBody(common.M{
			"gid":        "1",
			"trans_type": "msg",
		}).Get("http://localhost:8080/api/dtmsvr/abort")
		common.CheckRestySuccess(resp, err)
	})
	assert.Error(t, err)
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

func assertSucceed(t *testing.T, gid string) {
	WaitTransProcessed(gid)
	assert.Equal(t, "succeed", getTransStatus(gid))
}

func genMsg(gid string) *dtmcli.Msg {
	logrus.Printf("beginning a msg test ---------------- %s", gid)
	msg := dtmcli.NewMsg(examples.DtmServer, gid)
	msg.QueryPrepared = examples.Busi + "/CanSubmit"
	req := examples.GenTransReq(30, false, false)
	msg.Add(examples.Busi+"/TransOut", &req)
	msg.Add(examples.Busi+"/TransIn", &req)
	return msg
}

func genSaga(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	logrus.Printf("beginning a saga test ---------------- %s", gid)
	saga := dtmcli.NewSaga(examples.DtmServer, gid)
	req := examples.GenTransReq(30, outFailed, inFailed)
	saga.Add(examples.Busi+"/TransOut", examples.Busi+"/TransOutRevert", &req)
	saga.Add(examples.Busi+"/TransIn", examples.Busi+"/TransInRevert", &req)
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
	db := common.DbGet(config.DB)
	transInfo := &dtmcli.TransInfo{
		TransType:  "saga",
		Gid:        "gid2",
		BranchID:   "branch_id2",
		BranchType: "action",
	}
	db.Must().Exec("insert ignore into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type, reason) values('saga', 'gid1', 'branch_id1', 'action', 'saga')")
	_, err := dtmcli.ThroughBarrierCall(db.ToSQLDB(), transInfo, func(db *sql.Tx) (interface{}, error) {
		logrus.Printf("rollback gid2")
		return nil, fmt.Errorf("gid2 error")
	})
	asserts.Error(err, fmt.Errorf("gid2 error"))
	dbr := db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid1").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(0))
	gid2Res := common.M{"result": "first"}
	_, err = dtmcli.ThroughBarrierCall(db.ToSQLDB(), transInfo, func(db *sql.Tx) (interface{}, error) {
		logrus.Printf("submit gid2")
		return gid2Res, nil
	})
	asserts.Nil(err)
	dbr = db.Model(&dtmcli.BarrierModel{}).Where("gid=?", "gid2").Find(&[]dtmcli.BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	newResult, err := dtmcli.ThroughBarrierCall(db.ToSQLDB(), transInfo, func(db *sql.Tx) (interface{}, error) {
		logrus.Printf("submit gid2")
		return common.MS{"result": "ignored"}, nil
	})
	asserts.Equal(newResult, gid2Res)
}

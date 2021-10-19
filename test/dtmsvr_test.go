package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

var DtmServer = examples.DtmServer
var Busi = examples.Busi
var app *gin.Engine

// BarrierModel barrier model for gorm
type BarrierModel struct {
	common.ModelBase
	dtmcli.BranchBarrier
}

// TableName gorm table name
func (BarrierModel) TableName() string { return "dtm_barrier.barrier" }

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
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(gid))
}

func genMsg(gid string) *dtmcli.Msg {
	dtmcli.Logf("beginning a msg test ---------------- %s", gid)
	msg := dtmcli.NewMsg(examples.DtmServer, gid)
	msg.QueryPrepared = examples.Busi + "/CanSubmit"
	req := examples.GenTransReq(30, false, false)
	msg.Add(examples.Busi+"/TransOut", &req)
	msg.Add(examples.Busi+"/TransIn", &req)
	return msg
}

func transQuery(t *testing.T, gid string) {
	resp, err := dtmcli.RestyClient.R().SetQueryParam("gid", gid).Get(examples.DtmServer + "/query")
	e2p(err)
	m := M{}
	assert.Equal(t, resp.StatusCode(), 200)
	dtmcli.MustUnmarshalString(resp.String(), &m)
	assert.NotEqual(t, nil, m["transaction"])
	assert.Equal(t, 4, len(m["branches"].([]interface{})))

	resp, err = dtmcli.RestyClient.R().SetQueryParam("gid", "").Get(examples.DtmServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 500)

	resp, err = dtmcli.RestyClient.R().SetQueryParam("gid", "1").Get(examples.DtmServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 200)
	dtmcli.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, nil, m["transaction"])
	assert.Equal(t, 0, len(m["branches"].([]interface{})))

	resp, err = dtmcli.RestyClient.R().Get(examples.DtmServer + "/all")
	assert.Nil(t, err)
}

func TestSqlDB(t *testing.T) {
	asserts := assert.New(t)
	db := common.DbGet(config.DB)
	barrier := &dtmcli.BranchBarrier{
		TransType:  "saga",
		Gid:        "gid2",
		BranchID:   "branch_id2",
		BranchType: dtmcli.BranchAction,
	}
	db.Must().Exec("insert into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type, reason) values('saga', 'gid1', 'branch_id1', 'action', 'saga')")
	tx, err := db.ToSQLDB().Begin()
	asserts.Nil(err)
	err = barrier.Call(tx, func(db dtmcli.DB) error {
		dtmcli.Logf("rollback gid2")
		return fmt.Errorf("gid2 error")
	})
	asserts.Error(err, fmt.Errorf("gid2 error"))
	dbr := db.Model(&BarrierModel{}).Where("gid=?", "gid1").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	dbr = db.Model(&BarrierModel{}).Where("gid=?", "gid2").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(0))
	barrier.BarrierID = 0
	tx2, err := db.ToSQLDB().Begin()
	asserts.Nil(err)
	err = barrier.Call(tx2, func(db dtmcli.DB) error {
		dtmcli.Logf("submit gid2")
		return nil
	})
	asserts.Nil(err)
	dbr = db.Model(&BarrierModel{}).Where("gid=?", "gid2").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
}

func TestUpdateBranchAsync(t *testing.T) {
	common.DtmConfig.UpdateBranchSync = 0
	saga := genSaga("gid-update-branch-async", false, false)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Nil(t, err)
	WaitTransProcessed(saga.Gid)
	time.Sleep(dtmsvr.UpdateBranchAsyncInterval)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(saga.Gid))
	common.DtmConfig.UpdateBranchSync = 1
}

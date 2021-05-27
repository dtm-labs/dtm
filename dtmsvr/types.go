package dtmsvr

import (
	"fmt"
	"time"

	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
)

type M = map[string]interface{}

var p2e = common.P2E
var e2p = common.E2P

type TransGlobal struct {
	common.ModelBase
	Gid           string `json:"gid"`
	TransType     string `json:"trans_type"`
	Data          string `json:"data"`
	Status        string `json:"status"`
	QueryPrepared string `json:"query_prepared"`
	CommitTime    *time.Time
	FinishTime    *time.Time
	RollbackTime  *time.Time
}

func (*TransGlobal) TableName() string {
	return "trans_global"
}

func (t *TransGlobal) touch(db *common.MyDb) *gorm.DB {
	writeTransLog(t.Gid, "touch trans", "", "", "")
	return db.Model(&TransGlobal{}).Where("gid=?", t.Gid).Update("gid", t.Gid) // 更新update_time，避免被定时任务再次
}

func (t *TransGlobal) saveStatus(db *common.MyDb, status string) *gorm.DB {
	writeTransLog(t.Gid, "step change", status, "", "")
	dbr := db.Must().Model(t).Where("status=?", t.Status).Updates(M{
		"status":      status,
		"finish_time": time.Now(),
	})
	checkAffected(dbr)
	t.Status = status
	return dbr
}

type TransBranch struct {
	common.ModelBase
	Gid          string
	Url          string
	Data         string
	Branch       string
	BranchType   string
	Status       string
	FinishTime   *time.Time
	RollbackTime *time.Time
}

func (*TransBranch) TableName() string {
	return "trans_branch"
}

func (t *TransBranch) saveStatus(db *common.MyDb, status string) *gorm.DB {
	writeTransLog(t.Gid, "step change", status, t.Branch, "")
	dbr := db.Must().Model(t).Where("status=?", t.Status).Updates(M{
		"status":      status,
		"finish_time": time.Now(),
	})
	checkAffected(dbr)
	t.Status = status
	return dbr
}

func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("duplicate updating"))
	}
}

func (trans *TransGlobal) getProcessor() TransProcessor {
	if trans.TransType == "saga" {
		return &TransSagaProcessor{TransGlobal: trans}
	} else if trans.TransType == "tcc" {
		return &TransTccProcessor{TransGlobal: trans}
	} else if trans.TransType == "xa" {
		return &TransXaProcessor{TransGlobal: trans}
	}
	return nil
}

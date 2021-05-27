package dtmsvr

import (
	"fmt"
	"time"

	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
)

type M = map[string]interface{}

type TransGlobalModel struct {
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

func (*TransGlobalModel) TableName() string {
	return "trans_global"
}

func (t *TransGlobalModel) touch(db *common.MyDb) *gorm.DB {
	writeTransLog(t.Gid, "touch trans", "", "", "")
	return db.Model(&TransGlobalModel{}).Where("gid=?", t.Gid).Update("gid", t.Gid) // 更新update_time，避免被定时任务再次
}

func (t *TransGlobalModel) saveStatus(db *common.MyDb, status string) *gorm.DB {
	writeTransLog(t.Gid, "step change", status, "", "")
	dbr := db.Must().Model(t).Where("status=?", t.Status).Updates(M{
		"status":      status,
		"finish_time": time.Now(),
	})
	checkAffected(dbr)
	t.Status = status
	return dbr
}

type TransBranchModel struct {
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

func (*TransBranchModel) TableName() string {
	return "trans_branch"
}

func (t *TransBranchModel) saveStatus(db *common.MyDb, status string) *gorm.DB {
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

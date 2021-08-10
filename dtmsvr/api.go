package dtmsvr

import (
	"fmt"

	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm/clause"
)

func svcSubmit(t *TransGlobal, waitResult bool) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, t.Gid)
	if dbt != nil && dbt.Status != "prepared" && dbt.Status != "submitted" {
		return M{"dtm_result": "FAILURE", "message": fmt.Sprintf("current status %s, cannot sumbmit", dbt.Status)}, nil
	}
	t.Status = "submitted"
	t.saveNew(db)
	return t.Process(db, waitResult), nil

}

func svcPrepare(t *TransGlobal) (interface{}, error) {
	t.Status = "prepared"
	t.saveNew(dbGet())
	return dtmcli.ResultSuccess, nil
}

func svcAbort(t *TransGlobal, waitResult bool) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, t.Gid)
	if t.TransType != "xa" && t.TransType != "tcc" || dbt.Status != "prepared" && dbt.Status != "aborting" {
		return M{"dtm_result": "FAILURE", "message": fmt.Sprintf("trans type: %s current status %s, cannot abort", dbt.TransType, dbt.Status)}, nil
	}
	return dbt.Process(db, waitResult), nil
}

func svcRegisterTccBranch(branch *TransBranch, data dtmcli.MS) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, branch.Gid)
	if dbt.Status != "prepared" {
		return M{"dtm_result": "FAILURE", "message": fmt.Sprintf("current status: %s cannot register branch", dbt.Status)}, nil
	}

	branches := []TransBranch{*branch, *branch, *branch}
	for i, b := range []string{"cancel", "confirm", "try"} {
		branches[i].BranchType = b
		branches[i].URL = data[b]
	}

	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(dbGet(), config.TransCronInterval)
	return dtmcli.ResultSuccess, nil
}

func svcRegisterXaBranch(branch *TransBranch) (interface{}, error) {
	branch.Status = "prepared"
	db := dbGet()
	dbt := TransFromDb(db, branch.Gid)
	if dbt.Status != "prepared" {
		return M{"dtm_result": "FAILURE", "message": fmt.Sprintf("current status: %s cannot register branch", dbt.Status)}, nil
	}
	branches := []TransBranch{*branch, *branch}
	branches[0].BranchType = "rollback"
	branches[1].BranchType = "commit"
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(db, config.TransCronInterval)
	return dtmcli.ResultSuccess, nil
}

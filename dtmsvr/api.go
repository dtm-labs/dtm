package dtmsvr

import (
	"fmt"

	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm/clause"
)

func svcSubmit(t *TransGlobal) (interface{}, error) {
	db := dbGet()
	t.Status = dtmcli.StatusSubmitted
	err := t.saveNew(db)

	if err == errUniqueConflict {
		dbt := TransFromDb(db, t.Gid)
		if dbt.Status == dtmcli.StatusPrepared {
			updates := t.setNextCron(cronReset)
			db.Must().Model(t).Where("gid=? and status=?", t.Gid, dtmcli.StatusPrepared).Select(append(updates, "status")).Updates(t)
		} else if dbt.Status != dtmcli.StatusSubmitted {
			return M{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot sumbmit", dbt.Status)}, nil
		}
	}
	return t.Process(db), nil
}

func svcPrepare(t *TransGlobal) (interface{}, error) {
	t.Status = dtmcli.StatusPrepared
	err := t.saveNew(dbGet())
	if err == errUniqueConflict {
		dbt := TransFromDb(dbGet(), t.Gid)
		if dbt.Status != dtmcli.StatusPrepared {
			return M{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot prepare", dbt.Status)}, nil
		}
	}
	return dtmcli.MapSuccess, nil
}

func svcAbort(t *TransGlobal) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, t.Gid)
	if t.TransType != "xa" && t.TransType != "tcc" || dbt.Status != dtmcli.StatusPrepared && dbt.Status != dtmcli.StatusAborting {
		return M{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("trans type: '%s' current status '%s', cannot abort", dbt.TransType, dbt.Status)}, nil
	}
	dbt.changeStatus(db, dtmcli.StatusAborting)
	return dbt.Process(db), nil
}

func svcRegisterTccBranch(branch *TransBranch, data dtmcli.MS) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, branch.Gid)
	if dbt.Status != dtmcli.StatusPrepared {
		return M{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status: %s cannot register branch", dbt.Status)}, nil
	}

	branches := []TransBranch{*branch, *branch, *branch}
	for i, b := range []string{dtmcli.BranchCancel, dtmcli.BranchConfirm, dtmcli.BranchTry} {
		branches[i].BranchType = b
		branches[i].URL = data[b]
	}

	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(dbGet(), cronKeep)
	return dtmcli.MapSuccess, nil
}

func svcRegisterXaBranch(branch *TransBranch) (interface{}, error) {
	branch.Status = dtmcli.StatusPrepared
	db := dbGet()
	dbt := TransFromDb(db, branch.Gid)
	if dbt.Status != dtmcli.StatusPrepared {
		return M{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status: %s cannot register branch", dbt.Status)}, nil
	}
	branches := []TransBranch{*branch, *branch}
	branches[0].BranchType = dtmcli.BranchRollback
	branches[1].BranchType = dtmcli.BranchCommit
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(db, cronKeep)
	return dtmcli.MapSuccess, nil
}

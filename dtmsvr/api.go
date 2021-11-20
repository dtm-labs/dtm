/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"

	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func svcSubmit(t *TransGlobal) (interface{}, error) {
	db := dbGet()
	t.Status = dtmcli.StatusSubmitted
	err := t.saveNew(db)

	if err == errUniqueConflict {
		dbt := transFromDb(db.DB, t.Gid, false)
		if dbt.Status == dtmcli.StatusPrepared {
			updates := t.setNextCron(cronReset)
			dbr := db.Must().Model(&TransGlobal{}).Where("gid=? and status=?", t.Gid, dtmcli.StatusPrepared).Select(append(updates, "status")).Updates(t)
			checkAffected(dbr)
		} else if dbt.Status != dtmcli.StatusSubmitted {
			return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot sumbmit", dbt.Status)}, nil
		}
	}
	return t.Process(db), nil
}

func svcPrepare(t *TransGlobal) (interface{}, error) {
	t.Status = dtmcli.StatusPrepared
	err := t.saveNew(dbGet())
	if err == errUniqueConflict {
		dbt := transFromDb(dbGet().DB, t.Gid, false)
		if dbt.Status != dtmcli.StatusPrepared {
			return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot prepare", dbt.Status)}, nil
		}
	}
	return dtmcli.MapSuccess, nil
}

func svcAbort(t *TransGlobal) (interface{}, error) {
	db := dbGet()
	dbt := transFromDb(db.DB, t.Gid, false)
	if t.TransType != "xa" && t.TransType != "tcc" || dbt.Status != dtmcli.StatusPrepared && dbt.Status != dtmcli.StatusAborting {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("trans type: '%s' current status '%s', cannot abort", dbt.TransType, dbt.Status)}, nil
	}
	dbt.changeStatus(db, dtmcli.StatusAborting)
	return dbt.Process(db), nil
}

func svcRegisterBranch(branch *TransBranch, data map[string]string) (ret interface{}, rerr error) {
	err := dbGet().Transaction(func(db *gorm.DB) error {
		dbt := transFromDb(db, branch.Gid, true)
		if dbt.Status != dtmcli.StatusPrepared {
			ret = map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status: %s cannot register branch", dbt.Status)}
			return nil
		}

		branches := []TransBranch{*branch, *branch}
		if dbt.TransType == "tcc" {
			for i, b := range []string{dtmcli.BranchCancel, dtmcli.BranchConfirm} {
				branches[i].Op = b
				branches[i].URL = data[b]
			}
		} else if dbt.TransType == "xa" {
			branches[0].Op = dtmcli.BranchRollback
			branches[0].URL = data["url"]
			branches[1].Op = dtmcli.BranchCommit
			branches[1].URL = data["url"]
		} else {
			rerr = fmt.Errorf("unknow trans type: %s", dbt.TransType)
			return nil
		}

		dbr := db.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(branches)
		checkAffected(dbr)
		ret = dtmcli.MapSuccess
		return nil
	})
	e2p(err)
	return
}

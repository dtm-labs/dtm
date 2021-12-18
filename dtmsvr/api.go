/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmsvr/storage"
)

func svcSubmit(t *TransGlobal) (interface{}, error) {
	t.Status = dtmcli.StatusSubmitted
	err := t.saveNew()

	if err == storage.ErrUniqueConflict {
		dbt := GetTransGlobal(t.Gid)
		if dbt.Status == dtmcli.StatusPrepared {
			dbt.changeStatus(t.Status)
		} else if dbt.Status != dtmcli.StatusSubmitted {
			return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot sumbmit", dbt.Status)}, nil
		}
	}
	return t.Process(), nil
}

func svcPrepare(t *TransGlobal) (interface{}, error) {
	t.Status = dtmcli.StatusPrepared
	err := t.saveNew()
	if err == storage.ErrUniqueConflict {
		dbt := GetTransGlobal(t.Gid)
		if dbt.Status != dtmcli.StatusPrepared {
			return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("current status '%s', cannot prepare", dbt.Status)}, nil
		}
	}
	return dtmcli.MapSuccess, nil
}

func svcAbort(t *TransGlobal) (interface{}, error) {
	dbt := GetTransGlobal(t.Gid)
	if t.TransType != "xa" && t.TransType != "tcc" || dbt.Status != dtmcli.StatusPrepared && dbt.Status != dtmcli.StatusAborting {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": fmt.Sprintf("trans type: '%s' current status '%s', cannot abort", dbt.TransType, dbt.Status)}, nil
	}
	dbt.changeStatus(dtmcli.StatusAborting)
	return dbt.Process(), nil
}

func svcRegisterBranch(transType string, branch *TransBranch, data map[string]string) (ret interface{}, rerr error) {
	branches := []TransBranch{*branch, *branch}
	if transType == "tcc" {
		for i, b := range []string{dtmcli.BranchCancel, dtmcli.BranchConfirm} {
			branches[i].Op = b
			branches[i].URL = data[b]
		}
	} else if transType == "xa" {
		branches[0].Op = dtmcli.BranchRollback
		branches[0].URL = data["url"]
		branches[1].Op = dtmcli.BranchCommit
		branches[1].URL = data["url"]
	} else {
		return nil, fmt.Errorf("unknow trans type: %s", transType)
	}

	err := dtmimp.CatchP(func() {
		GetStore().LockGlobalSaveBranches(branch.Gid, dtmcli.StatusPrepared, branches, -1)
	})
	if err == storage.ErrNotFound {
		msg := fmt.Sprintf("no trans with gid: %s status: %s found", branch.Gid, dtmcli.StatusPrepared)
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": msg}, nil
	}
	return dtmimp.If(err != nil, nil, dtmcli.MapSuccess), err
}

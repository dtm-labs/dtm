/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
)

// Version store the passin version for dtm server
var Version = ""

func svcSubmit(t *TransGlobal) interface{} {
	t.Status = dtmcli.StatusSubmitted
	branches, err := t.saveNew()

	if err == storage.ErrUniqueConflict {
		dbt := GetTransGlobal(t.Gid)
		if dbt.Status == dtmcli.StatusPrepared {
			dbt.changeStatus(t.Status)
			branches = GetStore().FindBranches(t.Gid)
		} else if dbt.Status != dtmcli.StatusSubmitted {
			return fmt.Errorf("current status '%s', cannot sumbmit. %w", dbt.Status, dtmcli.ErrFailure)
		}
	}
	return t.Process(branches)
}

func svcPrepare(t *TransGlobal) interface{} {
	t.Status = dtmcli.StatusPrepared
	_, err := t.saveNew()
	if err == storage.ErrUniqueConflict {
		dbt := GetTransGlobal(t.Gid)
		if dbt.Status != dtmcli.StatusPrepared {
			return fmt.Errorf("current status '%s', cannot prepare. %w", dbt.Status, dtmcli.ErrFailure)
		}
		return nil
	}
	return err
}

func svcAbort(t *TransGlobal) interface{} {
	dbt := GetTransGlobal(t.Gid)
	if dbt.TransType == "msg" && dbt.Status == dtmcli.StatusPrepared {
		dbt.changeStatus(dtmcli.StatusFailed)
		return nil
	}
	if t.TransType != "xa" && t.TransType != "tcc" || dbt.Status != dtmcli.StatusPrepared && dbt.Status != dtmcli.StatusAborting {
		return fmt.Errorf("trans type: '%s' current status '%s', cannot abort. %w", dbt.TransType, dbt.Status, dtmcli.ErrFailure)
	}
	dbt.changeStatus(dtmcli.StatusAborting, withRollbackReason(t.RollbackReason))
	branches := GetStore().FindBranches(t.Gid)
	return dbt.Process(branches)
}

func svcForceStop(t *TransGlobal) interface{} {
	dbt := GetTransGlobal(t.Gid)
	if dbt.Status == dtmcli.StatusSucceed || dbt.Status == dtmcli.StatusFailed {
		return fmt.Errorf("global transaction force stop error. status: %s. error: %w", dbt.Status, dtmcli.ErrFailure)
	}
	dbt.changeStatus(dtmcli.StatusFailed)
	return nil
}

func svcRegisterBranch(transType string, branch *TransBranch, data map[string]string) error {
	branches := []TransBranch{*branch, *branch}
	if transType == "tcc" {
		for i, b := range []string{dtmimp.OpCancel, dtmimp.OpConfirm} {
			branches[i].Op = b
			branches[i].URL = data[b]
		}
	} else if transType == "xa" {
		branches[0].Op = dtmimp.OpRollback
		branches[0].URL = data["url"]
		branches[1].Op = dtmimp.OpCommit
		branches[1].URL = data["url"]
	} else {
		return fmt.Errorf("unknow trans type: %s", transType)
	}

	err := dtmimp.CatchP(func() {
		GetStore().LockGlobalSaveBranches(branch.Gid, dtmcli.StatusPrepared, branches, -1)
	})
	if err == storage.ErrNotFound {
		msg := fmt.Sprintf("no trans with gid: %s status: %s found", branch.Gid, dtmcli.StatusPrepared)
		logger.Errorf(msg)
		return fmt.Errorf("message: %s %w", msg, dtmcli.ErrFailure)
	}
	logger.Infof("LockGlobalSaveBranches result: %v: gid: %s old status: %s branches: %s",
		err, branch.Gid, dtmcli.StatusPrepared, dtmimp.MustMarshalString(branches))
	return err
}

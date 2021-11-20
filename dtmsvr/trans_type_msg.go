/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

type transMsgProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("msg", func(trans *TransGlobal) transProcessor { return &transMsgProcessor{TransGlobal: trans} })
}

func (t *transMsgProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	for i, step := range t.Steps {
		b := &TransBranch{
			Gid:      t.Gid,
			BranchID: fmt.Sprintf("%02d", i+1),
			BinData:  t.BinPayloads[i],
			URL:      step[dtmcli.BranchAction],
			Op:       dtmcli.BranchAction,
			Status:   dtmcli.StatusPrepared,
		}
		branches = append(branches, *b)
	}
	return branches
}

func (t *TransGlobal) mayQueryPrepared(db *common.DB) {
	if !t.needProcess() || t.Status == dtmcli.StatusSubmitted {
		return
	}
	body, err := t.getURLResult(t.QueryPrepared, "", "", nil)
	if strings.Contains(body, dtmcli.ResultSuccess) {
		t.changeStatus(db, dtmcli.StatusSubmitted)
	} else if strings.Contains(body, dtmcli.ResultFailure) {
		t.changeStatus(db, dtmcli.StatusFailed)
	} else if strings.Contains(body, dtmcli.ResultOngoing) {
		t.touch(db, cronReset)
	} else {
		dtmimp.LogRedf("getting result failed for %s. error: %s", t.QueryPrepared, err.Error())
		t.touch(db, cronBackoff)
	}
}

func (t *transMsgProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	t.mayQueryPrepared(db)
	if !t.needProcess() || t.Status == dtmcli.StatusPrepared {
		return nil
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.Op != dtmcli.BranchAction || branch.Status != dtmcli.StatusPrepared {
			continue
		}
		err := t.execBranch(db, branch)
		if err != nil {
			return err
		}
		if branch.Status != dtmcli.StatusSucceed {
			break
		}
	}
	if current == len(branches) { // msg 事务完成
		t.changeStatus(db, dtmcli.StatusSucceed)
		return nil
	}
	panic("msg go pass all branch")
}

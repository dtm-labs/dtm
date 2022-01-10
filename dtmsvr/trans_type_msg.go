/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"errors"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
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

func (t *TransGlobal) mayQueryPrepared() {
	if !t.needProcess() || t.Status == dtmcli.StatusSubmitted {
		return
	}
	err := t.getURLResult(t.QueryPrepared, "00", "msg", nil)
	if err == nil {
		t.changeStatus(dtmcli.StatusSubmitted)
	} else if errors.Is(err, dtmcli.ErrFailure) {
		t.changeStatus(dtmcli.StatusFailed)
	} else if errors.Is(err, dtmcli.ErrOngoing) {
		t.touchCronTime(cronReset)
	} else {
		logger.Errorf("getting result failed for %s. error: %v", t.QueryPrepared, err)
		t.touchCronTime(cronBackoff)
	}
}

func (t *transMsgProcessor) ProcessOnce(branches []TransBranch) error {
	t.mayQueryPrepared()
	if !t.needProcess() || t.Status == dtmcli.StatusPrepared {
		return nil
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.Op != dtmcli.BranchAction || branch.Status != dtmcli.StatusPrepared {
			continue
		}
		err := t.execBranch(branch, current)
		if err != nil {
			return err
		}
		if branch.Status != dtmcli.StatusSucceed {
			break
		}
	}
	if current == len(branches) { // msg 事务完成
		t.changeStatus(dtmcli.StatusSucceed)
		return nil
	}
	panic("msg go pass all branch")
}

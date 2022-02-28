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
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
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

type cMsgCustom struct {
	Delay uint64 //delay call branch, unit second
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
		t.touchCronTime(cronReset, 0)
	} else {
		logger.Errorf("getting result failed for %s. error: %v", t.QueryPrepared, err)
		t.touchCronTime(cronBackoff, 0)
	}
}

func (t *transMsgProcessor) ProcessOnce(branches []TransBranch) error {
	t.mayQueryPrepared()
	if !t.needProcess() || t.Status == dtmcli.StatusPrepared {
		return nil
	}
	cmc := cMsgCustom{Delay: 0}
	if t.CustomData != "" {
		dtmimp.MustUnmarshalString(t.CustomData, &cmc)
	}

	if cmc.Delay > 0 && t.needDelay(cmc.Delay) {
		t.touchCronTime(cronKeep, cmc.Delay)
		return nil
	}
	execBranch := func(current int) (bool, error) {
		branch := &branches[current]
		if branch.Op != dtmcli.BranchAction || branch.Status != dtmcli.StatusPrepared {
			return true, nil
		}
		err := t.execBranch(branch, current)
		if err != nil {
			if !errors.Is(err, dtmcli.ErrOngoing) {
				logger.Errorf("exec branch error: %v", err)
			}
			return false, err
		}
		if branch.Status != dtmcli.StatusSucceed {
			return false, nil
		}
		return true, nil
	}
	type branchResult struct {
		success bool
		err     error
	}
	waitChan := make(chan branchResult, len(branches))
	consumeWork := func(i int) error {
		success, err := execBranch(i)
		waitChan <- branchResult{
			success: success,
			err:     err,
		}
		return err
	}
	produceWork := func() {
		for i := 0; i < len(branches); i++ {
			if t.Concurrent {
				go func(i int) {
					_ = consumeWork(i)
				}(i)
				continue
			}
			err := consumeWork(i)
			if err != nil {
				return
			}
		}
	}
	go produceWork()
	successCnt := 0
	var err error
	for i := 0; i < len(branches); i++ {
		result := <-waitChan
		if result.err != nil {
			err = result.err
			if !t.Concurrent {
				return err
			}
		}
		if result.success {
			successCnt++
		}
	}
	if successCnt == len(branches) { // msg 事务完成
		t.changeStatus(dtmcli.StatusSucceed)
		return nil
	}
	panic("msg go pass all branch")
}

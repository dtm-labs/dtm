/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
)

type transXaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("xa", func(trans *TransGlobal) transProcessor { return &transXaProcessor{TransGlobal: trans} })
}

func (t *transXaProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *transXaProcessor) ProcessOnce(branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	if t.Status == dtmcli.StatusPrepared && t.isTimeout() {
		t.changeStatus(dtmcli.StatusAborting)
	}
	currentType := dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.BranchCommit, dtmcli.BranchRollback).(string)
	for i, branch := range branches {
		if branch.Op == currentType && branch.Status != dtmcli.StatusSucceed {
			err := t.execBranch(&branch, i)
			if err != nil {
				return err
			}
		}
	}
	t.changeStatus(dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

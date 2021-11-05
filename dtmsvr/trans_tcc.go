package dtmsvr

import (
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

type transTccProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("tcc", func(trans *TransGlobal) transProcessor { return &transTccProcessor{TransGlobal: trans} })
}

func (t *transTccProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *transTccProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	if t.Status == dtmcli.StatusPrepared && t.isTimeout() {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	branchType := dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.BranchConfirm, dtmcli.BranchCancel).(string)
	for current := len(branches) - 1; current >= 0; current-- {
		if branches[current].BranchType == branchType && branches[current].Status == dtmcli.StatusPrepared {
			dtmimp.Logf("branch info: current: %d ID: %d", current, branches[current].ID)
			err := t.execBranch(db, &branches[current])
			if err != nil {
				return err
			}
		}
	}
	t.changeStatus(db, dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

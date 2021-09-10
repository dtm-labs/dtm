package dtmsvr

import (
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
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

func (t *transTccProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == dtmcli.StatusSucceed || t.Status == dtmcli.StatusFailed {
		return
	}
	branchType := dtmcli.If(t.Status == dtmcli.StatusSubmitted, dtmcli.BranchConfirm, dtmcli.BranchCancel).(string)
	for current := len(branches) - 1; current >= 0; current-- {
		if branches[current].BranchType == branchType && branches[current].Status == dtmcli.StatusPrepared {
			t.execBranch(db, &branches[current])
		}
	}
	t.changeStatus(db, dtmcli.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
}

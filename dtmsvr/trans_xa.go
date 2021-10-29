package dtmsvr

import (
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
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

func (t *transXaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	currentType := dtmcli.If(t.Status == dtmcli.StatusSubmitted, dtmcli.BranchCommit, dtmcli.BranchRollback).(string)
	for _, branch := range branches {
		if branch.BranchType == currentType && branch.Status != dtmcli.StatusSucceed {
			err := t.execBranch(db, &branch)
			if err != nil {
				return err
			}
		}
	}
	t.changeStatus(db, dtmcli.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

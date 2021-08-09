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

func (t *transXaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == "succeed" {
		return
	}
	currentType := dtmcli.If(t.Status == "submitted", "commit", "rollback").(string)
	for _, branch := range branches {
		if branch.BranchType == currentType && branch.Status != "succeed" {
			t.execBranch(db, &branch)
		}
	}
	t.changeStatus(db, dtmcli.If(t.Status == "submitted", "succeed", "failed").(string))
}

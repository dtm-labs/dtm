package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
)

type TransXaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("xa", func(trans *TransGlobal) TransProcessor { return &TransXaProcessor{TransGlobal: trans} })
}

func (t *TransXaProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}
func (t *TransXaProcessor) ExecBranch(db *common.DB, branch *TransBranch) {
	resp, err := common.RestyClient.R().SetBody(M{
		"branch": branch.Branch,
		"action": common.If(t.Status == "prepared", "rollback", "commit"),
		"gid":    branch.Gid,
	}).Post(branch.Url)
	e2p(err)
	body := resp.String()
	t.touch(db)
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
	} else {
		panic(fmt.Errorf("bad response: %s", body))
	}
}

func (t *TransXaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == "succeed" {
		return
	}
	currentType := common.If(t.Status == "committed", "commit", "rollback").(string)
	for _, branch := range branches {
		if branch.BranchType == currentType && branch.Status != "succeed" {
			t.ExecBranch(db, &branch)
		}
	}
	t.changeStatus(db, common.If(t.Status == "committed", "succeed", "failed").(string))
}

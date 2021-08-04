package dtmsvr

import (
	"fmt"
	"strings"

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
func (t *transXaProcessor) ExecBranch(db *common.DB, branch *TransBranch) {
	resp, err := dtmcli.RestyClient.R().SetQueryParams(dtmcli.MS{
		"branch_id": branch.BranchID,
		"action":    dtmcli.If(t.Status == "prepared" || t.Status == "aborting", "rollback", "commit").(string),
		"gid":       branch.Gid,
	}).Post(branch.URL)
	e2p(err)
	body := resp.String()
	if strings.Contains(body, "SUCCESS") {
		t.touch(db, config.TransCronInterval)
		branch.changeStatus(db, "succeed")
	} else {
		panic(fmt.Errorf("bad response: %s", body))
	}
}

func (t *transXaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == "succeed" {
		return
	}
	currentType := dtmcli.If(t.Status == "submitted", "commit", "rollback").(string)
	for _, branch := range branches {
		if branch.BranchType == currentType && branch.Status != "succeed" {
			t.ExecBranch(db, &branch)
		}
	}
	t.changeStatus(db, dtmcli.If(t.Status == "submitted", "succeed", "failed").(string))
}

package dtmsvr

import (
	"fmt"
	"strings"

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

func (t *transTccProcessor) ExecBranch(db *common.DB, branch *TransBranch) {
	resp, err := dtmcli.RestyClient.R().SetBody(branch.Data).SetHeader("Content-type", "application/json").SetQueryParams(t.getBranchParams(branch)).Post(branch.URL)
	e2p(err)
	body := resp.String()
	if strings.Contains(body, "SUCCESS") {
		t.touch(db, config.TransCronInterval)
		branch.changeStatus(db, "succeed")
	} else {
		panic(fmt.Errorf("unknown response: %s, will be retried", body))
	}
}

func (t *transTccProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == "succeed" || t.Status == "failed" {
		return
	}
	branchType := dtmcli.If(t.Status == "submitted", "confirm", "cancel").(string)
	for current := len(branches) - 1; current >= 0; current-- {
		if branches[current].BranchType == branchType && branches[current].Status == "prepared" {
			t.ExecBranch(db, &branches[current])
		}
	}
	t.changeStatus(db, dtmcli.If(t.Status == "submitted", "succeed", "failed").(string))
}

package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
)

type TransTccProcessor struct {
	*TransGlobal
}

func (t *TransTccProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"cancel", "confirm", "try"} {
			branches = append(branches, TransBranch{
				Gid:        t.Gid,
				Branch:     fmt.Sprintf("%d", len(branches)+1),
				Data:       step["data"].(string),
				Url:        step[branchType].(string),
				BranchType: branchType,
				Status:     "prepared",
			})
		}
	}
	return branches
}

func (t *TransTccProcessor) ExecBranch(db *common.MyDb, branch *TransBranch) {
	resp, err := common.RestyClient.R().SetBody(branch.Data).SetQueryParam("gid", branch.Gid).Post(branch.Url)
	e2p(err)
	body := resp.String()
	t.touch(db)
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
	} else if branch.BranchType == "try" && strings.Contains(body, "FAIL") {
		branch.changeStatus(db, "failed")
	} else {
		panic(fmt.Errorf("unknown response: %s, will be retried", body))
	}
}

func (t *TransTccProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	t.MayQueryPrepared(db)
	if t.Status != "committed" {
		return
	}
	current := 0 // 当前正在处理的步骤
	// 先处理一轮正常try状态
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != "try" || branch.Status == "succeed" {
			continue
		}
		if branch.BranchType == "try" && branch.Status == "prepared" {
			t.ExecBranch(db, branch)
			if branch.Status != "succeed" {
				break
			}
		} else {
			break
		}
	}
	// 如果try全部成功，则处理confirm分支，否则处理cancel分支
	currentType := common.If(current == len(branches), "confirm", "cancel")
	for current--; current >= 0; current-- {
		branch := &branches[current]
		if branch.BranchType != currentType || branch.Status != "prepared" {
			continue
		}
		t.ExecBranch(db, branch)
	}
	t.changeStatus(db, common.If(currentType == "confirm", "succeed", "failed").(string))
}

package dtmsvr

import (
	"fmt"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

type transSagaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("saga", func(trans *TransGlobal) transProcessor { return &transSagaProcessor{TransGlobal: trans} })
}

func (t *transSagaProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	dtmcli.MustUnmarshalString(t.Data, &steps)
	for i, step := range steps {
		branch := fmt.Sprintf("%02d", i+1)
		for _, branchType := range []string{"compensate", "action"} {
			branches = append(branches, TransBranch{
				Gid:        t.Gid,
				BranchID:   branch,
				Data:       step["data"].(string),
				URL:        step[branchType].(string),
				BranchType: branchType,
				Status:     "prepared",
			})
		}
	}
	return branches
}

func (t *transSagaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	if t.Status == "failed" || t.Status == "succeed" {
		return
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != "action" || branch.Status == "succeed" {
			continue
		}
		// 找到了一个非succeed的action
		if branch.Status == "prepared" {
			t.execBranch(db, branch)
		}
		if branch.Status != "succeed" {
			break
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db, "succeed")
		return
	}
	if t.Status != "aborting" && t.Status != "failed" {
		t.changeStatus(db, "aborting")
	}
	for current = current - 1; current >= 0; current-- {
		branch := &branches[current]
		if branch.BranchType != "compensate" || branch.Status != "prepared" {
			continue
		}
		t.execBranch(db, branch)
	}
	t.changeStatus(db, "failed")
}

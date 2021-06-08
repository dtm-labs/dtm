package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
)

type TransSagaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("saga", func(trans *TransGlobal) TransProcessor { return &TransSagaProcessor{TransGlobal: trans} })
}

func (t *TransSagaProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"compensate", "action"} {
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

func (t *TransSagaProcessor) ExecBranch(db *common.MyDb, branch *TransBranch) {
	resp, err := common.RestyClient.R().SetBody(branch.Data).SetQueryParam("gid", branch.Gid).Post(branch.Url)
	e2p(err)
	body := resp.String()
	t.touch(db)
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
	} else if branch.BranchType == "action" && strings.Contains(body, "FAIL") {
		branch.changeStatus(db, "failed")
	} else {
		panic(fmt.Errorf("unknown response: %s, will be retried", body))
	}
}

func (t *TransSagaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	t.MayQueryPrepared(db)
	if t.Status != "committed" {
		return
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != "action" || branch.Status != "prepared" {
			continue
		}
		t.ExecBranch(db, branch)
		if branch.Status != "succeed" {
			break
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db, "succeed")
		return
	}
	for current = current - 1; current >= 0; current-- {
		branch := &branches[current]
		if branch.BranchType != "compensate" || branch.Status != "prepared" {
			continue
		}
		t.ExecBranch(db, branch)
	}
	if current != -1 {
		panic(fmt.Errorf("saga current not -1"))
	}
	t.changeStatus(db.Must(), "failed")
}

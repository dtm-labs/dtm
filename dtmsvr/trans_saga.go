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

func genSagaBranches(t *TransGlobal) []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	dtmcli.MustUnmarshalString(t.Data, &steps)
	for i, step := range steps {
		branch := fmt.Sprintf("%02d", i+1)
		for _, branchType := range []string{dtmcli.BranchCompensate, dtmcli.BranchAction} {
			branches = append(branches, TransBranch{
				Gid:        t.Gid,
				BranchID:   branch,
				Data:       step["data"].(string),
				URL:        step[branchType].(string),
				BranchType: branchType,
				Status:     dtmcli.StatusPrepared,
			})
		}
	}
	return branches
}

func (t *transSagaProcessor) GenBranches() []TransBranch {
	return genSagaBranches(t.TransGlobal)
}

func (t *transSagaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != dtmcli.BranchAction || branch.Status == dtmcli.StatusSucceed {
			continue
		}
		// 找到了一个非succeed的action
		if branch.Status == dtmcli.StatusPrepared {
			err := t.execBranch(db, branch)
			if err != nil {
				return err
			}
		}
		if branch.Status != dtmcli.StatusSucceed {
			break
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db, dtmcli.StatusSucceed)
		return nil
	}
	if t.Status != dtmcli.StatusAborting && t.Status != dtmcli.StatusFailed {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	for current = current - 1; current >= 0; current-- {
		branch := &branches[current]
		if branch.BranchType != dtmcli.BranchCompensate || branch.Status != dtmcli.StatusPrepared {
			continue
		}
		err := t.execBranch(db, branch)
		if err != nil {
			return err
		}
	}
	t.changeStatus(db, dtmcli.StatusFailed)
	return nil
}

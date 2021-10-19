package dtmsvr

import (
	"strings"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

type transMsgProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("msg", func(trans *TransGlobal) transProcessor { return &transMsgProcessor{TransGlobal: trans} })
}

func (t *transMsgProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	dtmcli.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		branches = append(branches, TransBranch{
			Gid:        t.Gid,
			BranchID:   GenGid(),
			Data:       step["data"].(string),
			URL:        step[dtmcli.BranchAction].(string),
			BranchType: dtmcli.BranchAction,
			Status:     dtmcli.StatusPrepared,
		})
	}
	return branches
}

func (t *TransGlobal) mayQueryPrepared(db *common.DB) {
	if t.Status != dtmcli.StatusPrepared {
		return
	}
	body := t.getURLResult(t.QueryPrepared, "", "", nil)
	if strings.Contains(body, dtmcli.ResultSuccess) {
		t.changeStatus(db, dtmcli.StatusSubmitted)
	} else if strings.Contains(body, dtmcli.ResultFailure) {
		t.changeStatus(db, dtmcli.StatusFailed)
	} else {
		t.touch(db, t.NextCronInterval*2)
	}
}

func (t *transMsgProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	t.mayQueryPrepared(db)
	if t.Status != dtmcli.StatusSubmitted {
		return
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != dtmcli.BranchAction || branch.Status != dtmcli.StatusPrepared {
			continue
		}
		t.execBranch(db, branch)
		if branch.Status != dtmcli.StatusSucceed {
			break
		}
	}
	if current == len(branches) { // msg 事务完成
		t.changeStatus(db, dtmcli.StatusSucceed)
		return
	}
	panic("msg go pass all branch")
}

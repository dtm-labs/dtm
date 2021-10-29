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
	if !t.needProcess() || t.Status == dtmcli.StatusSubmitted {
		return
	}
	body, err := t.getURLResult(t.QueryPrepared, "", "", nil)
	if strings.Contains(body, dtmcli.ResultSuccess) {
		t.changeStatus(db, dtmcli.StatusSubmitted)
	} else if strings.Contains(body, dtmcli.ResultFailure) {
		t.changeStatus(db, dtmcli.StatusFailed)
	} else if strings.Contains(body, dtmcli.ResultOngoing) {
		t.touch(db, cronReset)
	} else {
		dtmcli.LogRedf("getting result failed for %s. error: %s", t.QueryPrepared, err.Error())
		t.touch(db, cronBackoff)
	}
}

func (t *transMsgProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	t.mayQueryPrepared(db)
	if !t.needProcess() || t.Status == dtmcli.StatusPrepared {
		return nil
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		branch := &branches[current]
		if branch.BranchType != dtmcli.BranchAction || branch.Status != dtmcli.StatusPrepared {
			continue
		}
		err := t.execBranch(db, branch)
		if err != nil {
			return err
		}
		if branch.Status != dtmcli.StatusSucceed {
			break
		}
	}
	if current == len(branches) { // msg 事务完成
		t.changeStatus(db, dtmcli.StatusSucceed)
		return nil
	}
	panic("msg go pass all branch")
}

package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
)

type TransMsgProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("msg", func(trans *TransGlobal) TransProcessor { return &TransMsgProcessor{TransGlobal: trans} })
}

func (t *TransMsgProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		branches = append(branches, TransBranch{
			Gid:        t.Gid,
			Branch:     fmt.Sprintf("%d", len(branches)+1),
			Data:       step["data"].(string),
			Url:        step["action"].(string),
			BranchType: "action",
			Status:     "prepared",
		})
	}
	return branches
}

func (t *TransMsgProcessor) ExecBranch(db *common.MyDb, branch *TransBranch) {
	resp, err := common.RestyClient.R().SetBody(branch.Data).SetQueryParam("gid", branch.Gid).Post(branch.Url)
	e2p(err)
	body := resp.String()
	t.touch(db)
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
	} else {
		panic(fmt.Errorf("unknown response: %s, will be retried", body))
	}
}

func (t *TransMsgProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
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
	if current == len(branches) { // msg 事务完成
		t.changeStatus(db, "succeed")
		return
	}
	panic("msg go pass all branch")
}

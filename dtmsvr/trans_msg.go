package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
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
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		branches = append(branches, TransBranch{
			Gid:        t.Gid,
			BranchID:   GenGid(),
			Data:       step["data"].(string),
			URL:        step["action"].(string),
			BranchType: "action",
			Status:     "prepared",
		})
	}
	return branches
}

func (t *transMsgProcessor) ExecBranch(db *common.DB, branch *TransBranch) {
	resp, err := common.RestyClient.R().SetBody(branch.Data).SetQueryParams(t.getBranchParams(branch)).Post(branch.URL)
	e2p(err)
	body := resp.String()
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
		t.touch(db, config.TransCronInterval)
	} else {
		panic(fmt.Errorf("unknown response: %s, will be retried", body))
	}
}

func (t *TransGlobal) mayQueryPrepared(db *common.DB) {
	if t.Status != "prepared" {
		return
	}
	resp, err := common.RestyClient.R().SetQueryParam("gid", t.Gid).Get(t.QueryPrepared)
	e2p(err)
	body := resp.String()
	if strings.Contains(body, "SUCCESS") {
		t.changeStatus(db, "submitted")
	} else {
		t.touch(db, t.NextCronInterval*2)
	}
}

func (t *transMsgProcessor) ProcessOnce(db *common.DB, branches []TransBranch) {
	t.mayQueryPrepared(db)
	if t.Status != "submitted" {
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

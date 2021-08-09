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
			URL:        step["action"].(string),
			BranchType: "action",
			Status:     "prepared",
		})
	}
	return branches
}

func (t *TransGlobal) mayQueryPrepared(db *common.DB) {
	if t.Status != "prepared" {
		return
	}
	resp, err := dtmcli.RestyClient.R().SetQueryParam("gid", t.Gid).Get(t.QueryPrepared)
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
		t.execBranch(db, branch)
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

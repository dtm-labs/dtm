package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type TransProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(db *common.MyDb, branches []TransBranch)
	ExecBranch(db *common.MyDb, branch *TransBranch) string
}

type TransSagaProcessor struct {
	*TransGlobal
}

func (t *TransSagaProcessor) GenBranches() []TransBranch {
	nsteps := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"compensate", "action"} {
			nsteps = append(nsteps, TransBranch{
				Gid:        t.Gid,
				Branch:     fmt.Sprintf("%d", len(nsteps)+1),
				Data:       step["data"].(string),
				Url:        step[branchType].(string),
				BranchType: branchType,
				Status:     "prepared",
			})
		}
	}
	return nsteps
}

func (t *TransSagaProcessor) ExecBranch(db *common.MyDb, branche *TransBranch) string {
	return ""
}

func (t *TransSagaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	if t.Status == "prepared" {
		resp, err := common.RestyClient.R().SetQueryParam("gid", t.Gid).Get(t.QueryPrepared)
		e2p(err)
		body := resp.String()
		if strings.Contains(body, "FAIL") {
			preparedExpire := time.Now().Add(time.Duration(-config.PreparedExpire) * time.Second)
			logrus.Printf("create time: %s prepared expire: %s ", t.CreateTime.Local(), preparedExpire.Local())
			status := common.If(t.CreateTime.Before(preparedExpire), "canceled", "prepared").(string)
			if status != t.Status {
				t.changeStatus(db, status)
			}
			return
		} else if strings.Contains(body, "SUCCESS") {
			t.Status = "committed"
			t.SaveNew(db)
		} else {
			panic(fmt.Errorf("unknown result, will be retried: %s", body))
		}
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		step := branches[current]
		if step.BranchType == "compensate" && step.Status == "prepared" || step.BranchType == "action" && step.Status == "succeed" {
			continue
		}
		if step.BranchType == "action" && step.Status == "prepared" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			e2p(err)
			body := resp.String()

			t.touch(db.Must())
			if strings.Contains(body, "SUCCESS") {
				step.changeStatus(db.Must(), "succeed")
			} else if strings.Contains(body, "FAIL") {
				step.changeStatus(db.Must(), "failed")
				break
			} else {
				panic(fmt.Errorf("unknown response: %s, will be retried", body))
			}
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db.Must(), "succeed")
		return
	}
	for current = current - 1; current >= 0; current-- {
		step := branches[current]
		if step.BranchType != "compensate" || step.Status != "prepared" {
			continue
		}
		resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
		e2p(err)
		body := resp.String()
		if strings.Contains(body, "SUCCESS") {
			step.changeStatus(db.Must(), "failed")
		} else {
			panic(fmt.Errorf("expect compensate return SUCCESS"))
		}
	}
	if current != -1 {
		panic(fmt.Errorf("saga current not -1"))
	}
	t.changeStatus(db.Must(), "failed")
}

type TransTccProcessor struct {
	*TransGlobal
}

func (t *TransTccProcessor) GenBranches() []TransBranch {
	nsteps := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"cancel", "confirm", "try"} {
			nsteps = append(nsteps, TransBranch{
				Gid:        t.Gid,
				Branch:     fmt.Sprintf("%d", len(nsteps)+1),
				Data:       step["data"].(string),
				Url:        step[branchType].(string),
				BranchType: branchType,
				Status:     "prepared",
			})
		}
	}
	return nsteps
}

func (t *TransTccProcessor) ExecBranch(db *common.MyDb, branch *TransBranch) string {
	resp, err := common.RestyClient.R().SetBody(branch.Data).SetQueryParam("gid", branch.Gid).Post(branch.Url)
	e2p(err)
	body := resp.String()
	t.touch(db)
	if strings.Contains(body, "SUCCESS") {
		branch.changeStatus(db, "succeed")
		return "SUCCESS"
	}
	if branch.BranchType == "try" && strings.Contains(body, "FAIL") {
		branch.changeStatus(db, "failed")
		return "FAIL"
	}
	panic(fmt.Errorf("unknown response: %s, will be retried", body))
}

func (t *TransTccProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	current := 0 // 当前正在处理的步骤
	// 先处理一轮正常try状态
	for ; current < len(branches); current++ {
		step := &branches[current]
		if step.BranchType != "try" || step.Status == "succeed" {
			continue
		}
		if step.BranchType == "try" && step.Status == "prepared" {
			result := t.ExecBranch(db, step)
			if result == "FAIL" {
				break
			}
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

type TransXaProcessor struct {
	*TransGlobal
}

func (t *TransXaProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}
func (t *TransXaProcessor) ExecBranch(db *common.MyDb, branch *TransBranch) string {
	resp, err := common.RestyClient.R().SetBody(M{
		"branch": branch.Branch,
		"action": common.If(t.Status == "prepared", "rollback", "commit"),
		"gid":    branch.Gid,
	}).Post(branch.Url)
	e2p(err)
	body := resp.String()
	if !strings.Contains(body, "SUCCESS") {
		panic(fmt.Errorf("bad response: %s", body))
	}
	branch.changeStatus(db, "succeed")
	return "SUCCESS"
}

func (t *TransXaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	if t.Status == "succeed" {
		return
	}
	currentType := common.If(t.Status == "committed", "commit", "rollback").(string)
	for _, branch := range branches {
		if branch.BranchType == currentType && branch.Status != "succeed" {
			_ = t.ExecBranch(db, &branch)
			t.touch(db)
		}
	}
	t.changeStatus(db, common.If(t.Status == "committed", "succeed", "failed").(string))
}

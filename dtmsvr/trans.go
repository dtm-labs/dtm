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
		if step.BranchType == "compensate" && step.Status == "prepared" || step.BranchType == "action" && step.Status == "finished" {
			continue
		}
		if step.BranchType == "action" && step.Status == "prepared" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			e2p(err)
			body := resp.String()

			t.touch(db.Must())
			if strings.Contains(body, "SUCCESS") {
				step.changeStatus(db.Must(), "finished")
			} else if strings.Contains(body, "FAIL") {
				step.changeStatus(db.Must(), "rollbacked")
				break
			} else {
				panic(fmt.Errorf("unknown response: %s, will be retried", body))
			}
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db.Must(), "finished")
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
			step.changeStatus(db.Must(), "rollbacked")
		} else {
			panic(fmt.Errorf("expect compensate return SUCCESS"))
		}
	}
	if current != -1 {
		panic(fmt.Errorf("saga current not -1"))
	}
	t.changeStatus(db.Must(), "rollbacked")
}

type TransTccProcessor struct {
	*TransGlobal
}

func (t *TransTccProcessor) GenBranches() []TransBranch {
	nsteps := []TransBranch{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"rollback", "commit", "prepare"} {
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

func (t *TransTccProcessor) ExecBranch(db *common.MyDb, branche *TransBranch) string {
	return ""
}

func (t *TransTccProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		step := branches[current]
		if step.BranchType == "prepare" && step.Status == "finished" || step.BranchType != "commit" && step.Status == "prepared" {
			continue
		}
		if step.BranchType == "prepare" && step.Status == "prepared" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			e2p(err)
			body := resp.String()
			t.touch(db)
			if strings.Contains(body, "SUCCESS") {
				step.changeStatus(db, "finished")
			} else if strings.Contains(body, "FAIL") {
				step.changeStatus(db, "rollbacked")
				break
			} else {
				panic(fmt.Errorf("unknown response: %s, will be retried", body))
			}
		}
	}
	//////////////////////////////////////////////////
	if current == len(branches) { // tcc 事务完成
		t.changeStatus(db, "finished")
	}
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
	branch.changeStatus(db, common.If(t.Status == "prepared", "rollbacked", "finished").(string))
	return "SUCCESS"
}

func (t *TransXaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) {
	if t.Status == "finished" {
		return
	}
	if t.Status == "committed" {
		for _, branch := range branches {
			if branch.Status == "finished" {
				continue
			}
			_ = t.ExecBranch(db, &branch)
			t.touch(db) // 更新update_time，避免被定时任务再次
		}
		t.changeStatus(db, "finished")
	} else if t.Status == "prepared" { // 未commit直接处理的情况为回滚场景
		for _, branch := range branches {
			if branch.Status == "rollbacked" {
				continue
			}
			_ = t.ExecBranch(db, &branch)
			t.touch(db)
		}
		t.changeStatus(db, "rollbacked")
	} else {
		e2p(fmt.Errorf("bad trans status: %s", t.Status))
	}
}

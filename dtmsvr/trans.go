package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/yedf/dtm/common"
)

type TransProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(db *common.MyDb, branches []TransBranch) error
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

func (t *TransSagaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) error {
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		step := branches[current]
		if step.BranchType == "compensate" && step.Status == "prepared" || step.BranchType == "action" && step.Status == "finished" {
			continue
		}
		if step.BranchType == "action" && step.Status == "prepared" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			if err != nil {
				return err
			}
			body := resp.String()

			t.touch(db.Must())
			if strings.Contains(body, "SUCCESS") {
				step.changeStatus(db.Must(), "finished")
			} else if strings.Contains(body, "FAIL") {
				step.changeStatus(db.Must(), "rollbacked")
				break
			} else {
				return fmt.Errorf("unknown response: %s, will be retried", body)
			}
		}
	}
	if current == len(branches) { // saga 事务完成
		t.changeStatus(db.Must(), "finished")
		return nil
	}
	for current = current - 1; current >= 0; current-- {
		step := branches[current]
		if step.BranchType != "compensate" || step.Status != "prepared" {
			continue
		}
		resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
		if err != nil {
			return err
		}
		body := resp.String()
		if strings.Contains(body, "SUCCESS") {
			step.changeStatus(db.Must(), "rollbacked")
		} else {
			return fmt.Errorf("expect compensate return SUCCESS")
		}
	}
	if current != -1 {
		return fmt.Errorf("saga current not -1")
	}
	t.changeStatus(db.Must(), "rollbacked")
	return nil
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

func (t *TransTccProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) error {
	current := 0 // 当前正在处理的步骤
	for ; current < len(branches); current++ {
		step := branches[current]
		if step.BranchType == "prepare" && step.Status == "finished" || step.BranchType != "commit" && step.Status == "prepared" {
			continue
		}
		if step.BranchType == "prepare" && step.Status == "prepared" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			t.touch(db)
			if strings.Contains(body, "SUCCESS") {
				step.changeStatus(db, "finished")
			} else if strings.Contains(body, "FAIL") {
				step.changeStatus(db, "rollbacked")
				break
			} else {
				return fmt.Errorf("unknown response: %s, will be retried", body)
			}
		}
	}
	//////////////////////////////////////////////////
	if current == len(branches) { // tcc 事务完成
		t.changeStatus(db, "finished")
		return nil
	}
	return nil
}

type TransXaProcessor struct {
	*TransGlobal
}

func (t *TransXaProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *TransXaProcessor) ProcessOnce(db *common.MyDb, branches []TransBranch) error {
	gid := t.Gid
	if t.Status == "finished" {
		return nil
	}
	if t.Status == "committed" {
		for _, branch := range branches {
			if branch.Status == "finished" {
				continue
			}
			t.touch(db) // 更新update_time，避免被定时任务再次
			resp, err := common.RestyClient.R().SetBody(M{
				"branch": branch.Branch,
				"action": "commit",
				"gid":    branch.Gid,
			}).Post(branch.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			if !strings.Contains(body, "SUCCESS") {
				return fmt.Errorf("bad response: %s", body)
			}
			branch.changeStatus(db, "finished")
		}
		t.changeStatus(db, "finished")
	} else if t.Status == "prepared" { // 未commit直接处理的情况为回滚场景
		for _, branch := range branches {
			if branch.Status == "rollbacked" {
				continue
			}
			db.Must().Model(&TransGlobal{}).Where("gid=?", gid).Update("gid", gid) // 更新update_time，避免被定时任务再次
			resp, err := common.RestyClient.R().SetBody(M{
				"branch": branch.Branch,
				"action": "rollback",
				"gid":    branch.Gid,
			}).Post(branch.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			if !strings.Contains(body, "SUCCESS") {
				return fmt.Errorf("bad response: %s", body)
			}
			branch.changeStatus(db, "rollbacked")
		}
		t.changeStatus(db, "rollbacked")
	} else {
		return fmt.Errorf("bad trans status: %s", t.Status)
	}
	return nil
}

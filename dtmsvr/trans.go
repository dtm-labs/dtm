package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/yedf/dtm/common"
)

type Trans interface {
	GetDataBranches() []TransBranchModel
	ProcessOnce(db *common.MyDb, branches []TransBranchModel) error
}

func GetTrans(trans *TransGlobalModel) Trans {
	if trans.TransType == "saga" {
		return &TransSaga{TransGlobalModel: trans}
	} else if trans.TransType == "tcc" {
		return &TransTcc{TransGlobalModel: trans}
	} else if trans.TransType == "xa" {
		return &TransXa{TransGlobalModel: trans}
	}
	return nil
}

type TransSaga struct {
	*TransGlobalModel
}

func (t *TransSaga) GetDataBranches() []TransBranchModel {
	nsteps := []TransBranchModel{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"compensate", "action"} {
			nsteps = append(nsteps, TransBranchModel{
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

func (t *TransSaga) ProcessOnce(db *common.MyDb, branches []TransBranchModel) error {
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
				step.saveStatus(db.Must(), "finished")
			} else if strings.Contains(body, "FAIL") {
				step.saveStatus(db.Must(), "rollbacked")
				break
			} else {
				return fmt.Errorf("unknown response: %s, will be retried", body)
			}
		}
	}
	if current == len(branches) { // saga 事务完成
		t.saveStatus(db.Must(), "finished")
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
			step.saveStatus(db.Must(), "rollbacked")
		} else {
			return fmt.Errorf("expect compensate return SUCCESS")
		}
	}
	if current != -1 {
		return fmt.Errorf("saga current not -1")
	}
	t.saveStatus(db.Must(), "rollbacked")
	return nil
}

type TransTcc struct {
	*TransGlobalModel
}

func (t *TransTcc) GetDataBranches() []TransBranchModel {
	nsteps := []TransBranchModel{}
	steps := []M{}
	common.MustUnmarshalString(t.Data, &steps)
	for _, step := range steps {
		for _, branchType := range []string{"rollback", "commit", "prepare"} {
			nsteps = append(nsteps, TransBranchModel{
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

func (t *TransTcc) ProcessOnce(db *common.MyDb, branches []TransBranchModel) error {
	gid := t.Gid
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
			db.Must().Model(&TransGlobalModel{}).Where("gid=?", gid).Update("gid", gid) // 更新update_time，避免被定时任务再次
			if strings.Contains(body, "SUCCESS") {
				writeTransLog(gid, "step finished", "finished", step.Branch, "")
				dbr := db.Must().Model(&step).Where("status=?", "prepared").Updates(M{
					"status":      "finished",
					"finish_time": time.Now(),
				})
				checkAffected(dbr)
			} else if strings.Contains(body, "FAIL") {
				writeTransLog(gid, "step rollbacked", "rollbacked", step.Branch, "")
				dbr := db.Must().Model(&step).Where("status=?", "prepared").Updates(M{
					"status":        "rollbacked",
					"rollback_time": time.Now(),
				})
				checkAffected(dbr)
				break
			} else {
				return fmt.Errorf("unknown response: %s, will be retried", body)
			}
		}
	}
	//////////////////////////////////////////////////
	if current == len(branches) { // tcc 事务完成
		writeTransLog(gid, "saga finished", "finished", "", "")
		dbr := db.Must().Model(&TransGlobalModel{}).Where("gid=? and status=?", gid, "committed").Updates(M{
			"status":      "finished",
			"finish_time": time.Now(),
		})
		checkAffected(dbr)
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
			writeTransLog(gid, "step rollbacked", "rollbacked", step.Branch, "")
			dbr := db.Must().Model(&step).Where("status=?", step.Status).Updates(M{
				"status":        "rollbacked",
				"rollback_time": time.Now(),
			})
			checkAffected(dbr)
		} else {
			return fmt.Errorf("expect compensate return SUCCESS")
		}
	}
	if current != -1 {
		return fmt.Errorf("saga current not -1")
	}
	writeTransLog(gid, "saga rollbacked", "rollbacked", "", "")
	dbr := db.Must().Model(&TransGlobalModel{}).Where("status=? and gid=?", "committed", gid).Updates(M{
		"status":        "rollbacked",
		"rollback_time": time.Now(),
	})
	checkAffected(dbr)
	return nil
}

type TransXa struct {
	*TransGlobalModel
}

func (t *TransXa) GetDataBranches() []TransBranchModel {
	return []TransBranchModel{}
}

func (t *TransXa) ProcessOnce(db *common.MyDb, branches []TransBranchModel) error {
	gid := t.Gid
	if t.Status == "finished" {
		return nil
	}
	if t.Status == "committed" {
		for _, branch := range branches {
			if branch.Status == "finished" {
				continue
			}
			db.Must().Model(&TransGlobalModel{}).Where("gid=?", gid).Update("gid", gid) // 更新update_time，避免被定时任务再次
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
			writeTransLog(gid, "step finished", "finished", branch.Branch, "")
			db.Must().Model(&branch).Where("status=?", "prepared").Updates(M{
				"status":      "finished",
				"finish_time": time.Now(),
			})
		}
		writeTransLog(gid, "xa finished", "finished", "", "")
		db.Must().Model(&TransGlobalModel{}).Where("gid=? and status=?", gid, "committed").Updates(M{
			"status":      "finished",
			"finish_time": time.Now(),
		})
	} else if t.Status == "prepared" { // 未commit直接处理的情况为回滚场景
		for _, branch := range branches {
			if branch.Status == "rollbacked" {
				continue
			}
			db.Must().Model(&TransGlobalModel{}).Where("gid=?", gid).Update("gid", gid) // 更新update_time，避免被定时任务再次
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
			writeTransLog(gid, "step rollbacked", "rollbacked", branch.Branch, "")
			db.Must().Model(&branch).Where("status=?", "prepared").Updates(M{
				"status":      "rollbacked",
				"finish_time": time.Now(),
			})
		}
		writeTransLog(gid, "xa rollbacked", "rollbacked", "", "")
		db.Must().Model(&TransGlobalModel{}).Where("gid=? and status=?", gid, "prepared").Updates(M{
			"status":      "rollbacked",
			"finish_time": time.Now(),
		})
	} else {
		return fmt.Errorf("bad trans status: %s", t.Status)
	}
	return nil
}

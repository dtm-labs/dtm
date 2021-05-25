package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func saveCommitted(m *TransGlobalModel) {
	db := dbGet()
	m.Status = "committed"
	err := db.Transaction(func(db1 *gorm.DB) error {
		db := &common.MyDb{DB: db1}
		writeTransLog(m.Gid, "save committed", m.Status, "", m.Data)
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&m)
		if dbr.RowsAffected == 0 {
			writeTransLog(m.Gid, "change status", m.Status, "", "")
			db.Must().Model(&m).Where("status=?", "prepared").Update("status", "committed")
		}
		if m.TransType == "saga" {
			nsteps := []TransBranchModel{}
			steps := []M{}
			common.MustUnmarshalString(m.Data, &steps)
			for _, step := range steps {
				nsteps = append(nsteps, TransBranchModel{
					Gid:        m.Gid,
					Branch:     fmt.Sprintf("%d", len(nsteps)+1),
					Data:       step["data"].(string),
					Url:        step["compensate"].(string),
					BranchType: "compensate",
					Status:     "prepared",
				})
				nsteps = append(nsteps, TransBranchModel{
					Gid:        m.Gid,
					Branch:     fmt.Sprintf("%d", len(nsteps)+1),
					Data:       step["data"].(string),
					Url:        step["action"].(string),
					BranchType: "action",
					Status:     "prepared",
				})
			}
			writeTransLog(m.Gid, "save steps", m.Status, "", common.MustMarshalString(nsteps))
			db.Must().Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&nsteps)
		}
		return nil
	})
	common.PanicIfError(err)
}

var TransProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitTransProcessed(gid string) {
	id := <-TransProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-TransProcessedTestChan
	}
}

func ProcessTrans(trans *TransGlobalModel) {
	err := innerProcessTrans(trans)
	if err != nil {
		logrus.Errorf("process trans ignore error: %s", err.Error())
	}
	if TransProcessedTestChan != nil {
		TransProcessedTestChan <- trans.Gid
	}
}
func innerProcessTrans(trans *TransGlobalModel) (rerr error) {
	branches := []TransBranchModel{}
	db := dbGet()
	db.Must().Where("gid=?", trans.Gid).Order("id asc").Find(&branches)
	if trans.TransType == "saga" {
		return innerProcessCommittedSaga(trans, db, branches)
	} else if trans.TransType == "xa" {
		return innerProcessCommittedXa(trans, db, branches)
	}
	panic(fmt.Errorf("unkown trans type: %s", trans.TransType))
}

func innerProcessCommittedXa(trans *TransGlobalModel, db *common.MyDb, branches []TransBranchModel) error {
	gid := trans.Gid
	if trans.Status == "finished" {
		return nil
	}
	if trans.Status == "committed" {
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
	} else if trans.Status == "prepared" { // 未commit直接处理的情况为回滚场景
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
		return fmt.Errorf("bad trans status: %s", trans.Status)
	}
	return nil
}

func innerProcessCommittedSaga(trans *TransGlobalModel, db *common.MyDb, branches []TransBranchModel) error {
	gid := trans.Gid
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
	if current == len(branches) { // saga 事务完成
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

func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("duplicate updating"))
	}
}

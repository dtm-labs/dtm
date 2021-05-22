package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func saveCommitedSagaModel(m *SagaModel) {
	db := DbGet()
	m.Status = "commited"
	err := db.Transaction(func(db1 *gorm.DB) error {
		db := &MyDb{DB: db1}
		writeTransLog(m.Gid, "save commited", m.Status, -1, m.Steps)
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&m)
		if dbr.RowsAffected == 0 {
			writeTransLog(m.Gid, "change status", m.Status, -1, "")
			db.Must().Model(&m).Where("status=?", "prepared").Update("status", "commited")
		}
		nsteps := []SagaStepModel{}
		steps := []M{}
		common.MustUnmarshalString(m.Steps, &steps)
		for _, step := range steps {
			nsteps = append(nsteps, SagaStepModel{
				Gid:    m.Gid,
				Step:   len(nsteps) + 1,
				Data:   step["post_data"].(string),
				Url:    step["compensate"].(string),
				Type:   "compensate",
				Status: "pending",
			})
			nsteps = append(nsteps, SagaStepModel{
				Gid:    m.Gid,
				Step:   len(nsteps) + 1,
				Data:   step["post_data"].(string),
				Url:    step["action"].(string),
				Type:   "action",
				Status: "pending",
			})
		}
		writeTransLog(m.Gid, "save steps", m.Status, -1, common.MustMarshalString(nsteps))
		db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&nsteps)
		return nil
	})
	common.PanicIfError(err)
}

var SagaProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitCommitedSaga(gid string) {
	id := <-SagaProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-SagaProcessedTestChan
	}
}

func ProcessCommitedSaga(gid string) {
	err := innerProcessCommitedSaga(gid)
	if err != nil {
		logrus.Errorf("process commited saga error: %s", err.Error())
	}
	if SagaProcessedTestChan != nil {
		SagaProcessedTestChan <- gid
	}
}
func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("duplicate updating"))
	}
}

func innerProcessCommitedSaga(gid string) (rerr error) {
	steps := []SagaStepModel{}
	db := DbGet()
	db.Must().Order("id asc").Find(&steps)
	current := 0 // 当前正在处理的步骤
	for ; current < len(steps); current++ {
		step := steps[current]
		if step.Type == "compensate" && step.Status == "pending" || step.Type == "action" && step.Status == "finished" {
			continue
		}
		if step.Type == "action" && step.Status == "pending" {
			resp, err := dtm.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			db.Must().Model(&SagaModel{}).Where("gid=?", gid).Update("gid", gid) // 更新update_time，避免被定时任务再次
			if strings.Contains(body, "SUCCESS") {
				writeTransLog(gid, "step finished", "finished", step.Step, "")
				dbr := db.Must().Model(&step).Where("status=?", "pending").Updates(M{
					"status":      "finished",
					"finish_time": time.Now(),
				})
				checkAffected(dbr)
			} else if strings.Contains(body, "FAIL") {
				writeTransLog(gid, "step rollbacked", "rollbacked", step.Step, "")
				dbr := db.Must().Model(&step).Where("status=?", "pending").Updates(M{
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
	if current == len(steps) { // saga 事务完成
		writeTransLog(gid, "saga finished", "finished", -1, "")
		dbr := db.Must().Model(&SagaModel{}).Where("gid=? and status=?", gid, "commited").Updates(M{
			"status":      "finished",
			"finish_time": time.Now(),
		})
		checkAffected(dbr)
		return nil
	}
	for current = current - 1; current >= 0; current-- {
		step := steps[current]
		if step.Type != "compensate" || step.Status != "pending" {
			continue
		}
		resp, err := dtm.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
		if err != nil {
			return err
		}
		body := resp.String()
		if strings.Contains(body, "SUCCESS") {
			writeTransLog(gid, "step rollbacked", "rollbacked", step.Step, "")
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
	writeTransLog(gid, "saga rollbacked", "rollbacked", -1, "")
	dbr := db.Must().Model(&SagaModel{}).Where("status=? and gid=?", "commited", gid).Updates(M{
		"status":        "rollbacked",
		"rollback_time": time.Now(),
	})
	checkAffected(dbr)
	return nil
}

package dtmsvr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddRoute(engine *gin.Engine) {
	engine.POST("/api/dtmsvr/prepare", Prepare)
	engine.POST("/api/dtmsvr/commit", Commit)
}

func getSagaModelFromContext(c *gin.Context) *SagaModel {
	data := M{}
	b, err := c.GetRawData()
	common.PanicIfError(err)
	common.MustUnmarshal(b, &data)
	logrus.Printf("creating saga model in prepare")
	data["steps"] = common.MustMarshalString(data["steps"])
	m := SagaModel{}
	common.MustRemarshal(data, &m)
	return &m
}

func Prepare(c *gin.Context) {
	db := DbGet()
	m := getSagaModelFromContext(c)
	m.Status = "prepared"
	writeTransLog(m.Gid, "save prepared", m.Status, -1, m.Steps)
	db1 := db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&m)
	common.PanicIfError(db1.Error)
	c.JSON(200, M{"message": "SUCCESS"})
}

func Commit(c *gin.Context) {
	m := getSagaModelFromContext(c)
	saveCommitedSagaModel(m)
	go ProcessCommitedSaga(m.Gid)
	c.JSON(200, M{"message": "SUCCESS"})
}

func saveCommitedSagaModel(m *SagaModel) {
	db := DbGet()
	m.Status = "commited"
	stepInserted := false
	err := db.Transaction(func(db *gorm.DB) error {
		writeTransLog(m.Gid, "save commited", m.Status, -1, m.Steps)
		dbr := db.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&m)
		if dbr.Error == nil && dbr.RowsAffected == 0 {
			writeTransLog(m.Gid, "change status", m.Status, -1, "")
			dbr = db.Model(&m).Where("status=?", "prepared").Update("status", "commited")
		}
		common.PanicIfError(dbr.Error)
		nsteps := []SagaStepModel{}
		steps := []M{}
		err := json.Unmarshal([]byte(m.Steps), &steps)
		common.PanicIfError(err)
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
		r := db.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&nsteps)
		if db.Error != nil {
			return db.Error
		}
		if r.RowsAffected == int64(len(nsteps)) {
			stepInserted = true
		}
		logrus.Printf("rows affected: %d nsteps length: %d, stepInersted: %t", r.RowsAffected, int64(len(nsteps)), stepInserted)
		return db.Error
	})
	common.PanicIfError(err)
	if !stepInserted {
		return
	}
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

func innerProcessCommitedSaga(gid string) (rerr error) {
	steps := []SagaStepModel{}
	db := DbGet()
	db1 := db.Order("id asc").Find(&steps)
	if db1.Error != nil {
		return db1.Error
	}
	checkAffected := func(db1 *gorm.DB) {
		common.PanicIfError(db1.Error)
		if db1.RowsAffected == 0 {
			panic(fmt.Errorf("duplicate updating"))
		}
	}
	current := 0 // 当前正在处理的步骤
	for ; current < len(steps); current++ {
		step := steps[current]
		if step.Type == "compensate" && step.Status == "pending" || step.Type == "action" && step.Status == "finished" {
			continue
		}
		if step.Type == "action" && step.Status == "pending" {
			resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			if strings.Contains(body, "SUCCESS") {
				writeTransLog(gid, "step finished", "finished", step.Step, "")
				dbr := db.Model(&step).Where("status=?", "pending").Updates(M{
					"status":      "finished",
					"finish_time": time.Now(),
				})
				checkAffected(dbr)
			} else if strings.Contains(body, "FAIL") {
				writeTransLog(gid, "step rollbacked", "rollbacked", step.Step, "")
				dbr := db.Model(&step).Where("status=?", "pending").Updates(M{
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
		dbr := db.Model(&SagaModel{}).Where("gid=? and status=?", gid, "commited").Updates(M{
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
		resp, err := common.RestyClient.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
		if err != nil {
			return err
		}
		body := resp.String()
		if strings.Contains(body, "SUCCESS") {
			writeTransLog(gid, "step rollbacked", "rollbacked", step.Step, "")
			dbr := db.Model(&step).Where("status=?", step.Status).Updates(M{
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
	dbr := db.Model(&SagaModel{}).Where("status=? and gid=?", "commited", gid).Updates(M{
		"status":        "rollbacked",
		"rollback_time": time.Now(),
	})
	checkAffected(dbr)
	return nil
}

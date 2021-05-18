package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ModelBase struct {
	ID         uint
	CreateTime time.Time `gorm:"autoCreateTime"`
	UpdateTime time.Time `gorm:"autoUpdateTime"`
}
type SagaModel struct {
	ModelBase
	Gid          string
	Steps        string
	TransQuery   string
	Status       string
	FinishTime   time.Time
	RollbackTime time.Time
}

func (*SagaModel) TableName() string {
	return "test1.a_saga"
}

type SagaStepModel struct {
	ModelBase
	Gid          string
	Data         string
	Step         int
	Url          string
	Type         string
	Status       string
	FinishTime   string
	RollbackTime string
}

func (*SagaStepModel) TableName() string {
	return "test1.a_saga_step"
}

func handlePreparedMsg(data gin.H) {
	db := DbGet()
	logrus.Printf("creating saga model in prepare")
	data["steps"] = common.MustMarshalString(data["steps"])
	m := SagaModel{}
	err := common.Map2Obj(data, &m)
	common.PanicIfError(err)
	m.Status = "prepared"
	db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&m)
}

func handleCommitedMsg(data gin.H) {
	db := DbGet()
	logrus.Printf("creating saga model in commited")
	steps := data["steps"].([]interface{})
	data["steps"] = common.MustMarshalString(data["steps"])
	m := SagaModel{}
	err := common.Map2Obj(data, &m)
	common.PanicIfError(err)
	m.Status = "processing"
	stepInserted := false
	err = db.Transaction(func(db *gorm.DB) error {
		db.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&m)
		if db.Error == nil && db.RowsAffected == 0 {
			db.Model(&m).Where("status=?", "prepared").Update("status", "processing")
		}
		nsteps := []SagaStepModel{}
		for _, step1 := range steps {
			step := step1.(map[string]interface{})
			nsteps = append(nsteps, SagaStepModel{
				Gid:    m.Gid,
				Step:   len(nsteps) + 1,
				Data:   common.MustMarshalString(step["post_data"]),
				Url:    step["compensate"].(string),
				Type:   "compensate",
				Status: "pending",
			})
			nsteps = append(nsteps, SagaStepModel{
				Gid:    m.Gid,
				Step:   len(nsteps) + 1,
				Data:   common.MustMarshalString(step["post_data"]),
				Url:    step["action"].(string),
				Type:   "action",
				Status: "pending",
			})
		}
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
	err = ProcessCommitedSaga(m.Gid)
	if err != nil {
		logrus.Printf("---------------handle commited msmg error: %s", err.Error())
	}
}

func ProcessCommitedSaga(gid string) (rerr error) {
	steps := []SagaStepModel{}
	db := DbGet()
	db1 := db.Order("id asc").Find(&steps)
	if db1.Error != nil {
		return db1.Error
	}
	current := 0 // 当前正在处理的步骤
	tx := []*gorm.DB{db.Begin()}
	defer func() { // 如果直接return出去，则rollback当前的事务
		tx[0].Rollback()
		if err := recover(); err != nil {
			rerr = err.(error)
		}
	}()
	checkAndCommit := func(db1 *gorm.DB) {
		common.PanicIfError(db1.Error)
		if db1.RowsAffected == 0 {
			panic(fmt.Errorf("duplicate updating"))
		}
		dbr := tx[0].Commit()
		common.PanicIfError(dbr.Error)
		tx[0] = db.Begin()
		common.PanicIfError(tx[0].Error)
	}
	for ; current < len(steps); current++ {
		step := steps[current]
		if step.Type == "compensate" && step.Status == "pending" || step.Type == "action" && step.Status == "finished" {
			continue
		}
		if step.Type == "action" && step.Status == "pending" {
			resp, err := client.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
			if err != nil {
				return err
			}
			body := resp.String()
			if strings.Contains(body, "SUCCESS") {
				dbr := tx[0].Model(&step).Where("status=?", "pending").Updates(M{
					"status":      "finished",
					"finish_time": time.Now(),
				})
				checkAndCommit(dbr)
			} else if strings.Contains(body, "FAIL") {
				dbr := tx[0].Model(&step).Where("status=?", "pending").Updates(M{
					"status":        "rollbacked",
					"rollback_time": time.Now(),
				})
				checkAndCommit(dbr)
			}
		}
	}
	if current == len(steps) { // saga 事务完成
		dbr := tx[0].Model(&SagaModel{}).Where("gid=? and status=?", gid, "processing").Updates(M{
			"status":      "finished",
			"finish_time": time.Now(),
		})
		checkAndCommit(dbr)
		return nil
	}
	for current = len(steps) - 1; current >= 0; current-- {
		step := steps[current]
		if step.Type != "compensate" || step.Status != "pending" {
			continue
		}
		resp, err := client.R().SetBody(step.Data).SetQueryParam("gid", step.Gid).Post(step.Url)
		if err != nil {
			return err
		}
		body := resp.String()
		if strings.Contains(body, "SUCCESS") {
			dbr := tx[0].Model(&step).Where("status=?", step.Status).Updates(M{
				"status":        "rollbacked",
				"rollback_time": time.Now(),
			})
			checkAndCommit(dbr)
		} else {
			return fmt.Errorf("expect compensate return SUCCESS")
		}
	}
	if current != -1 {
		return fmt.Errorf("saga current not -1")
	}
	dbr := tx[0].Model(&SagaModel{}).Where("status=?", "processing").Updates(M{
		"status":        "rollbacked",
		"rollback_time": time.Now(),
	})
	checkAndCommit(dbr)
	return nil
}

func StartConsumePreparedMsg(consumers int) {
	logrus.Printf("start to consume prepared msg")
	r := RabbitmqGet()
	for i := 0; i < consumers; i++ {
		go func() {
			que := r.QueueNew(RabbitmqConstPrepared)
			que.WaitAndHandle(handlePreparedMsg)
		}()
	}
}

func StartConsumeCommitedMsg(consumers int) {
	logrus.Printf("start to consume commited msg")
	r := RabbitmqGet()
	for i := 0; i < consumers; i++ {
		go func() {
			que := r.QueueNew(RabbitmqConstCommited)
			que.WaitAndHandle(handleCommitedMsg)
		}()
	}
}

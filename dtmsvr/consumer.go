package dtmsvr

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
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
	FinishTime   string
	RollbackTime string
}

func (*SagaModel) TableName() string {
	return "test1.a_saga"
}

func handlePreparedMsg(data gin.H) {
	data["gid"] = "4eHhkCxVsQ1"
	db := DbGet()
	// db.Model(&SagaModel{}).Clauses(clause.OnConflict{
	// 	DoNothing: true,
	// }).Create(data)

	logrus.Printf("creating saga model")
	db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&SagaModel{
		Gid:        data["gid"].(string),
		Steps:      string(common.MustMarshal(data["steps"])),
		TransQuery: data["trans_query"].(string),
		Status:     "prepared",
	})
}

func handleCommitedMsg(data gin.H) {

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

package dtmsvr

import (
	"time"
)

type M = map[string]interface{}

type ModelBase struct {
	ID         uint
	CreateTime time.Time `gorm:"autoCreateTime"`
	UpdateTime time.Time `gorm:"autoUpdateTime"`
}
type SagaModel struct {
	ModelBase
	Gid          string `json:"gid"`
	Steps        string `json:"steps"`
	TransQuery   string `json:"trans_query"`
	Status       string `json:"status"`
	FinishTime   time.Time
	RollbackTime time.Time
}

func (*SagaModel) TableName() string {
	return "saga"
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
	return "saga_step"
}

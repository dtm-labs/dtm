package dtmcli

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// Saga struct of saga
type Saga struct {
	SagaData
	Server string
}

// SagaData sage data
type SagaData struct {
	Gid       string     `json:"gid"`
	TransType string     `json:"trans_type"`
	Steps     []SagaStep `json:"steps"`
}

// SagaStep one step of saga
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	Data       string `json:"data"`
}

// NewSaga create a saga
func NewSaga(server string, gid string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid:       gid,
			TransType: "saga",
		},
		Server: server,
	}
}

// Add add a saga step
func (s *Saga) Add(action string, compensate string, postData interface{}) *Saga {
	logrus.Printf("saga %s Add %s %s %v", s.Gid, action, compensate, postData)
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
		Data:       common.MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Submit submit the saga trans
func (s *Saga) Submit() error {
	_, err := s.SubmitExt(&TransOptions{})
	return err
}

// SubmitExt 高级submit，更多的选项和更详细的返回值
func (s *Saga) SubmitExt(opt *TransOptions) (TransStatus, error) {
	return callDtm(s.Server, &s.SagaData, "submit", opt)
}

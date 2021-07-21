package dtmcli

import (
	"fmt"
	"strings"

	jsonitor "github.com/json-iterator/go"
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
func NewSaga(server string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid:       GenGid(server),
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
	logrus.Printf("committing %s body: %v", s.Gid, &s.SagaData)
	resp, err := common.RestyClient.R().SetBody(&s.SagaData).Post(fmt.Sprintf("%s/submit", s.Server))
	if err != nil {
		return err
	}
	if !strings.Contains(resp.String(), "SUCCESS") {
		return fmt.Errorf("submit failed: %v", resp.Body())
	}
	s.Gid = jsonitor.Get(resp.Body(), "gid").ToString()
	return nil
}

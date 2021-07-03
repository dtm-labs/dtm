package dtmcli

import (
	"fmt"

	jsonitor "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type Saga struct {
	SagaData
	Server string
}

type SagaData struct {
	Gid       string     `json:"gid"`
	TransType string     `json:"trans_type"`
	Steps     []SagaStep `json:"steps"`
}
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	Data       string `json:"data"`
}

func NewSaga(server string) *Saga {
	return &Saga{
		SagaData: SagaData{
			TransType: "saga",
		},
		Server: server,
	}
}
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

func (s *Saga) Submit() error {
	logrus.Printf("committing %s body: %v", s.Gid, &s.SagaData)
	resp, err := common.RestyClient.R().SetBody(&s.SagaData).Post(fmt.Sprintf("%s/submit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("submit failed: %v", resp.Body())
	}
	s.Gid = jsonitor.Get(resp.Body(), "gid").ToString()
	return nil
}

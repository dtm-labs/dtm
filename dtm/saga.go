package dtm

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type Saga struct {
	SagaData
	Server string
}

type SagaData struct {
	Gid        string     `json:"gid"`
	Steps      []SagaStep `json:"steps"`
	TransQuery string     `json:"trans_query"`
}
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	PostData   string `json:"post_data"`
}

func SagaNew(server string, gid string, transQuery string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid:        gid,
			TransQuery: transQuery,
		},
		Server: server,
	}
}
func (s *Saga) Add(action string, compensate string, postData interface{}) error {
	logrus.Printf("saga %s Add %s %s %v", s.Gid, action, compensate, postData)
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
		PostData:   common.MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *Saga) getBody() *SagaData {
	return &s.SagaData
}

func (s *Saga) Prepare() error {
	logrus.Printf("preparing %s body: %v", s.Gid, s.getBody())
	resp, err := common.RestyClient.R().SetBody(s.getBody()).Post(fmt.Sprintf("%s/prepare", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("prepare failed: %v", resp.Body())
	}
	return nil
}

func (s *Saga) Commit() error {
	logrus.Printf("committing %s body: %v", s.Gid, s.getBody())
	resp, err := common.RestyClient.R().SetBody(s.getBody()).Post(fmt.Sprintf("%s/commit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("commit failed: %v", resp.Body())
	}
	return nil
}

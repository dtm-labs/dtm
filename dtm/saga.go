package dtm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type SagaData struct {
	Gid        string     `json:"gid"`
	Steps      []SagaStep `json:"steps"`
	TransQuery string     `json:"trans_query"`
}
type Saga struct {
	SagaData
	Server string
}
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	PostData   gin.H  `json:"post_data"`
}

func SagaNew(server string, gid string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid: gid,
		},
		Server: server,
	}
}
func (s *Saga) Add(action string, compensate string, postData gin.H) error {
	logrus.Printf("saga %s Add %s %s %v", s.Gid, action, compensate, postData)
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
		PostData:   postData,
	}
	step.PostData = postData
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *Saga) getBody() *SagaData {
	return &s.SagaData
}

func (s *Saga) Prepare(url string) error {
	s.TransQuery = url
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

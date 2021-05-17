package dtm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var client *resty.Client = resty.New()

type Saga struct {
	Server     string     `json:"server"`
	Gid        string     `json:"gid"`
	Steps      []SagaStep `json:"steps"`
	TransQuery string     `json:"trans_query"`
}
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	PostData   gin.H  `json:"post_data"`
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

func (s *Saga) getBody() gin.H {
	return gin.H{
		"gid":         s.Gid,
		"trans_query": s.TransQuery,
		"steps":       s.Steps,
	}
}

func (s *Saga) Prepare(url string) error {
	s.TransQuery = url
	logrus.Printf("preparing %s body: %v", s.Gid, s.getBody())
	resp, err := client.R().SetBody(s.getBody()).Post(fmt.Sprintf("%s/prepare", s.Server))
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
	resp, err := client.R().SetBody(s.getBody()).Post(fmt.Sprintf("%s/commit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("commit failed: %v", resp.Body())
	}
	return nil
}

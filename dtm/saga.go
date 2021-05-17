package dtm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var client *resty.Client = resty.New()

type Saga struct {
	Server     string
	Gid        string
	Steps      []SagaStep
	TransQuery string
}
type SagaStep struct {
	Action     string
	Compensate string
	PostData   interface{}
}

func (s *Saga) Add(action string, compensate string, postData interface{}) error {
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
	}
	step.PostData = postData
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *Saga) Prepare(url string) error {
	s.TransQuery = url
	resp, err := client.R().SetBody(gin.H{
		"Gid":        s.Gid,
		"TransQuery": s.TransQuery,
		"Steps":      s.Steps,
	}).Post(fmt.Sprintf("%s/prepare", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("prepare failed: %v", resp.Body())
	}
	return nil
}

func (s *Saga) Commit() error {
	resp, err := client.R().SetBody(gin.H{}).Post(fmt.Sprintf("%s/commit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("commit failed: %v", resp.Body())
	}
	return nil
}

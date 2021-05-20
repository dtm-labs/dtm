package dtm

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
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
	d, err := json.Marshal(postData)
	if err != nil {
		return err
	}
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
		PostData:   string(d),
	}
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *Saga) Prepare() error {
	logrus.Printf("preparing %s body: %v", s.Gid, &s.SagaData)
	resp, err := RestyClient.R().SetBody(&s.SagaData).Post(fmt.Sprintf("%s/prepare", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("prepare failed: %v", resp.Body())
	}
	return nil
}

func (s *Saga) Commit() error {
	logrus.Printf("committing %s body: %v", s.Gid, &s.SagaData)
	resp, err := RestyClient.R().SetBody(&s.SagaData).Post(fmt.Sprintf("%s/commit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("commit failed: %v", resp.Body())
	}
	return nil
}

// 辅助工具与代码
var RestyClient = resty.New()

func init() {
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logrus.Printf("requesting: %s %s %v", r.Method, r.URL, r.Body)
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logrus.Printf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}

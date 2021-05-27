package dtm

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type Tcc struct {
	TccData
	Server string
}

type TccData struct {
	Gid           string    `json:"gid"`
	TransType     string    `json:"trans_type"`
	Steps         []TccStep `json:"steps"`
	QueryPrepared string    `json:"query_prepared"`
}
type TccStep struct {
	Prepare  string `json:"prepare"`
	Commit   string `json:"commit"`
	Rollback string `json:"rollback"`
	Data     string `json:"data"`
}

func TccNew(server string, gid string, queryPrepared string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid:           gid,
			TransType:     "tcc",
			QueryPrepared: queryPrepared,
		},
		Server: server,
	}
}
func (s *Tcc) Add(prepare string, commit string, rollback string, data interface{}) error {
	logrus.Printf("tcc %s Add %s %s %s %v", s.Gid, prepare, commit, rollback, data)
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	step := TccStep{
		Prepare:  prepare,
		Commit:   commit,
		Rollback: rollback,
		Data:     string(d),
	}
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *Tcc) Prepare() error {
	logrus.Printf("preparing %s body: %v", s.Gid, &s.TccData)
	resp, err := common.RestyClient.R().SetBody(&s.TccData).Post(fmt.Sprintf("%s/prepare", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("prepare failed: %v", resp.Body())
	}
	return nil
}

func (s *Tcc) Commit() error {
	logrus.Printf("committing %s body: %v", s.Gid, &s.TccData)
	resp, err := common.RestyClient.R().SetBody(&s.TccData).Post(fmt.Sprintf("%s/commit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("commit failed: %v", resp.Body())
	}
	return nil
}

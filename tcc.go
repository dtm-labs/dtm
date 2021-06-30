package dtm

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type Tcc struct {
	TccData
	Server string
}

type TccData struct {
	Gid       string    `json:"gid"`
	TransType string    `json:"trans_type"`
	Steps     []TccStep `json:"steps"`
}
type TccStep struct {
	Try     string `json:"try"`
	Confirm string `json:"confirm"`
	Cancel  string `json:"cancel"`
	Data    string `json:"data"`
}

func TccNew(server string, gid string) *Tcc {
	return &Tcc{
		TccData: TccData{
			Gid:       gid,
			TransType: "tcc",
		},
		Server: server,
	}
}
func (s *Tcc) Add(try string, confirm string, cancel string, data interface{}) *Tcc {
	logrus.Printf("tcc %s Add %s %s %s %v", s.Gid, try, confirm, cancel, data)
	step := TccStep{
		Try:     try,
		Confirm: confirm,
		Cancel:  cancel,
		Data:    common.MustMarshalString(data),
	}
	s.Steps = append(s.Steps, step)
	return s
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

package dtmcli

import (
	"fmt"

	jsonitor "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type Msg struct {
	MsgData
	Server string
}

type MsgData struct {
	Gid           string    `json:"gid"`
	TransType     string    `json:"trans_type"`
	Steps         []MsgStep `json:"steps"`
	QueryPrepared string    `json:"query_prepared"`
}
type MsgStep struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func NewMsg(server string) *Msg {
	return &Msg{
		MsgData: MsgData{
			Gid:       common.GenGid(),
			TransType: "msg",
		},
		Server: server,
	}
}
func (s *Msg) Add(action string, postData interface{}) *Msg {
	logrus.Printf("msg %s Add %s %v", s.Gid, action, postData)
	step := MsgStep{
		Action: action,
		Data:   common.MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

func (s *Msg) Submit() error {
	logrus.Printf("committing %s body: %v", s.Gid, &s.MsgData)
	resp, err := common.RestyClient.R().SetBody(&s.MsgData).Post(fmt.Sprintf("%s/submit", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("submit failed: %v", resp.Body())
	}
	s.Gid = jsonitor.Get(resp.Body(), "gid").ToString()
	return nil
}

func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = common.OrString(queryPrepared, s.QueryPrepared)
	logrus.Printf("preparing %s body: %v", s.Gid, &s.MsgData)
	resp, err := common.RestyClient.R().SetBody(&s.MsgData).Post(fmt.Sprintf("%s/prepare", s.Server))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("prepare failed: %v", resp.Body())
	}
	return nil
}

package dtmcli

import (
	"fmt"

	jsonitor "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// Msg reliable msg type
type Msg struct {
	MsgData
	Server string
}

// MsgData msg data
type MsgData struct {
	Gid           string    `json:"gid"`
	TransType     string    `json:"trans_type"`
	Steps         []MsgStep `json:"steps"`
	QueryPrepared string    `json:"query_prepared"`
}

// MsgStep struct of one step msg
type MsgStep struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

// NewMsg create new msg
func NewMsg(server string, gid string) *Msg {
	return &Msg{
		MsgData: MsgData{
			Gid:       gid,
			TransType: "msg",
		},
		Server: server,
	}
}

// Add add a new step
func (s *Msg) Add(action string, postData interface{}) *Msg {
	logrus.Printf("msg %s Add %s %v", s.Gid, action, postData)
	step := MsgStep{
		Action: action,
		Data:   common.MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Submit submit the msg
func (s *Msg) Submit() error {
	logrus.Printf("committing %s body: %v", s.Gid, &s.MsgData)
	resp, err := common.RestyClient.R().SetBody(&s.MsgData).Post(fmt.Sprintf("%s/submit", s.Server))
	rerr := CheckDtmResponse(resp, err)
	if rerr != nil {
		return rerr
	}
	s.Gid = jsonitor.Get(resp.Body(), "gid").ToString()
	return nil
}

// Prepare prepare the msg
func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = common.OrString(queryPrepared, s.QueryPrepared)
	logrus.Printf("preparing %s body: %v", s.Gid, &s.MsgData)
	resp, err := common.RestyClient.R().SetBody(&s.MsgData).Post(fmt.Sprintf("%s/prepare", s.Server))
	rerr := CheckDtmResponse(resp, err)
	if rerr != nil {
		return rerr
	}
	return nil
}

package dtmcli

import (
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

// Prepare prepare the msg
func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = common.OrString(queryPrepared, s.QueryPrepared)
	return callDtmSimple(s.Server, &s.MsgData, "prepare")
}

// Submit submit the msg
func (s *Msg) Submit() error {
	return callDtmSimple(s.Server, &s.MsgData, "submit")
}

// SubmitExt 高级submit，更多的选项和更详细的返回值
func (s *Msg) SubmitExt(opt *TransOptions) (TransStatus, error) {
	return CallDtm(s.Server, &s.MsgData, "submit", opt)
}

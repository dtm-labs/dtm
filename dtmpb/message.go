package dtmpb

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
)

// MsgGrpc reliable msg type
type MsgGrpc struct {
	MsgDataGrpc
	dtmcli.TransBase
}

// MsgDataGrpc msg data
type MsgDataGrpc struct {
	dtmcli.TransData
	Steps         []MsgStepGrpc `json:"steps"`
	QueryPrepared string        `json:"query_prepared"`
}

// MsgStepGrpc struct of one step msg
type MsgStepGrpc struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

// NewMsgGrpc create new msg
func NewMsgGrpc(server string, gid string) *MsgGrpc {
	return &MsgGrpc{
		MsgDataGrpc: MsgDataGrpc{TransData: dtmcli.TransData{
			Gid:       gid,
			TransType: "msg",
		}},
		TransBase: dtmcli.TransBase{
			Dtm: server,
		},
	}
}

// Add add a new step
func (s *MsgGrpc) Add(action string, data []byte) *MsgGrpc {
	dtmcli.Logf("msg %s Add %s %v", s.MsgDataGrpc.Gid, action, string(data))
	step := MsgStepGrpc{
		Action: action,
		Data:   string(data),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Submit submit the msg
func (s *MsgGrpc) Submit() error {
	_, err := MustGetDtmClient(s.Dtm).Submit(context.Background(), &DtmRequest{
		Gid:       s.Gid,
		TransType: s.TransType,
		Data:      dtmcli.MustMarshalString(&s.Steps),
	})
	return err
}

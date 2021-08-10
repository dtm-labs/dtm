package dtmgrpc

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
)

// MsgGrpc reliable msg type
type MsgGrpc struct {
	dtmcli.MsgData
	dtmcli.TransBase
}

// NewMsgGrpc create new msg
func NewMsgGrpc(server string, gid string) *MsgGrpc {
	return &MsgGrpc{
		MsgData: dtmcli.MsgData{TransData: dtmcli.TransData{
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
	step := dtmcli.MsgStep{
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

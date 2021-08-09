package dtmcli

// MsgGrpc reliable msg type
type MsgGrpc struct {
	MsgDataGrpc
	TransBase
}

// MsgDataGrpc msg data
type MsgDataGrpc struct {
	TransData
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
		MsgDataGrpc: MsgDataGrpc{TransData: TransData{
			Gid:       gid,
			TransType: "msg",
		}},
		TransBase: TransBase{
			Dtm: server,
		},
	}
}

// Add add a new step
func (s *MsgGrpc) Add(action string, postData interface{}) *MsgGrpc {
	Logf("msg %s Add %s %v", s.MsgDataGrpc.Gid, action, postData)
	step := MsgStepGrpc{
		Action: action,
		Data:   MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Submit submit the msg
func (s *MsgGrpc) Submit() error {
	return s.CallDtm(&s.MsgDataGrpc, "submit")
}

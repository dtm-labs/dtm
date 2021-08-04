package dtmcli

// Msg reliable msg type
type Msg struct {
	MsgData
	TransBase
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
		TransBase: TransBase{
			Dtm: server,
		},
	}
}

// Add add a new step
func (s *Msg) Add(action string, postData interface{}) *Msg {
	Logf("msg %s Add %s %v", s.MsgData.Gid, action, postData)
	step := MsgStep{
		Action: action,
		Data:   MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Prepare prepare the msg
func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = OrString(queryPrepared, s.QueryPrepared)
	return s.CallDtm(&s.MsgData, "prepare")
}

// Submit submit the msg
func (s *Msg) Submit() error {
	return s.CallDtm(&s.MsgData, "submit")
}

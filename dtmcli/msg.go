package dtmcli

import "github.com/yedf/dtm/dtmcli/dtmimp"

// Msg reliable msg type
type Msg struct {
	dtmimp.TransBase
}

// NewMsg create new msg
func NewMsg(server string, gid string) *Msg {
	return &Msg{TransBase: *dtmimp.NewTransBase(gid, "msg", server, "")}
}

// Add add a new step
func (s *Msg) Add(action string, postData interface{}) *Msg {
	s.Steps = append(s.Steps, map[string]string{"action": action})
	s.Payloads = append(s.Payloads, dtmimp.MustMarshalString(postData))
	return s
}

// Prepare prepare the msg, msg will later be submitted
func (s *Msg) Prepare(queryPrepared string) error {
	s.QueryPrepared = dtmimp.OrString(queryPrepared, s.QueryPrepared)
	return dtmimp.TransCallDtm(&s.TransBase, s, "prepare")
}

// Submit submit the msg
func (s *Msg) Submit() error {
	return dtmimp.TransCallDtm(&s.TransBase, s, "submit")
}

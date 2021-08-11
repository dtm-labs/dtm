package dtmcli

// Saga struct of saga
type Saga struct {
	TransBase
	Steps []SagaStep `json:"steps"`
}

// SagaStep one step of saga
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	Data       string `json:"data"`
}

// NewSaga create a saga
func NewSaga(server string, gid string) *Saga {
	return &Saga{TransBase: *NewTransBase(gid, "saga", server, "")}
}

// Add add a saga step
func (s *Saga) Add(action string, compensate string, postData interface{}) *Saga {
	s.Steps = append(s.Steps, SagaStep{
		Action:     action,
		Compensate: compensate,
		Data:       MustMarshalString(postData),
	})
	return s
}

// Submit submit the saga trans
func (s *Saga) Submit() error {
	return s.callDtm(s, "submit")
}

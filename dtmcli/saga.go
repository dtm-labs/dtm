package dtmcli

// Saga struct of saga
type Saga struct {
	SagaData
	TransBase
}

// SagaData sage data
type SagaData struct {
	Gid       string     `json:"gid"`
	TransType string     `json:"trans_type"`
	Steps     []SagaStep `json:"steps"`
}

// SagaStep one step of saga
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	Data       string `json:"data"`
}

// NewSaga create a saga
func NewSaga(server string, gid string) *Saga {
	return &Saga{
		SagaData: SagaData{
			Gid:       gid,
			TransType: "saga",
		},
		TransBase: TransBase{
			Dtm: server,
		},
	}
}

// Add add a saga step
func (s *Saga) Add(action string, compensate string, postData interface{}) *Saga {
	Logf("saga %s Add %s %s %v", s.SagaData.Gid, action, compensate, postData)
	step := SagaStep{
		Action:     action,
		Compensate: compensate,
		Data:       MustMarshalString(postData),
	}
	s.Steps = append(s.Steps, step)
	return s
}

// Submit submit the saga trans
func (s *Saga) Submit() error {
	return s.CallDtm(&s.SagaData, "submit")
}

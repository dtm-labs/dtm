package dtmgrpc

import (
	context "context"

	"github.com/yedf/dtm/dtmcli"
)

// SagaGrpc struct of saga
type SagaGrpc struct {
	dtmcli.TransBase
	Steps []dtmcli.SagaStep `json:"steps"`
}

// NewSaga create a saga
func NewSaga(server string, gid string) *SagaGrpc {
	return &SagaGrpc{TransBase: *dtmcli.NewTransBase(gid, "saga", server, "")}
}

// Add add a saga step
func (s *SagaGrpc) Add(action string, compensate string, busiData []byte) *SagaGrpc {
	s.Steps = append(s.Steps, dtmcli.SagaStep{
		Action:     action,
		Compensate: compensate,
		Data:       string(busiData),
	})
	return s
}

// Submit submit the saga trans
func (s *SagaGrpc) Submit() error {
	_, err := MustGetDtmClient(s.Dtm).Submit(context.Background(), &DtmRequest{
		Gid:       s.Gid,
		TransType: s.TransType,
		Data:      dtmcli.MustMarshalString(&s.Steps),
	})
	return err
}

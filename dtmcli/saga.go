package dtmcli

import "fmt"

// Saga struct of saga
type Saga struct {
	TransBase
	Steps      []SagaStep `json:"steps"`
	orders     map[int][]int
	concurrent bool
}

// SagaStep one step of saga
type SagaStep struct {
	Action     string `json:"action"`
	Compensate string `json:"compensate"`
	Data       string `json:"data"`
}

// NewSaga create a saga
func NewSaga(server string, gid string) *Saga {
	return &Saga{TransBase: *NewTransBase(gid, "saga", server, ""), orders: map[int][]int{}}
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

// AddStepOrder specify that step should be after preSteps. Step is larger than all the element in preSteps
func (s *Saga) AddStepOrder(step int, preSteps []int) *Saga {
	PanicIf(step > len(s.Steps), fmt.Errorf("step value: %d is invalid. which cannot be larger than total steps: %d", step, len(s.Steps)))
	s.orders[step] = preSteps
	return s
}

// EnableConcurrent enable the concurrent exec of sub trans
func (s *Saga) EnableConcurrent() *Saga {
	s.concurrent = true
	return s
}

// Submit submit the saga trans
func (s *Saga) Submit() error {
	if s.concurrent {
		s.CustomData = MustMarshalString(M{"orders": s.orders, "concurrent": s.concurrent})
	}
	return s.callDtm(s, "submit")
}

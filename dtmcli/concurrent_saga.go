package dtmcli

import "fmt"

// ConcurrentSaga struct of concurrent saga
type ConcurrentSaga struct {
	Saga
	orders map[int][]int
}

// NewConcurrentSaga create a concurrent saga
func NewConcurrentSaga(server string, gid string) *ConcurrentSaga {
	return &ConcurrentSaga{Saga: Saga{TransBase: *NewTransBase(gid, "csaga", server, "")}, orders: map[int][]int{}}
}

// AddStepOrder specify that step should be after preSteps. Step is larger than all the element in preSteps
func (s *ConcurrentSaga) AddStepOrder(step int, preSteps []int) *ConcurrentSaga {
	PanicIf(step > len(s.Steps), fmt.Errorf("step value: %d is invalid. which cannot be larger than total steps: %d", step, len(s.Steps)))
	s.orders[step] = preSteps
	return s
}

// Submit submit the saga trans
func (s *ConcurrentSaga) Submit() error {
	if len(s.orders) > 0 {
		s.CustomData = MustMarshalString(M{"orders": s.orders})
	}
	return s.callDtm(s, "submit")
}

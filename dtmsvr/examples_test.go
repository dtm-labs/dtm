package dtmsvr

import (
	"testing"

	"github.com/yedf/dtm/examples"
)

func TestExamples(t *testing.T) {
	// for coverage
	examples.QsStartSvr()
	assertSucceed(t, examples.QsFireRequest())
	assertSucceed(t, examples.MsgFireRequest())
	assertSucceed(t, examples.SagaBarrierFireRequest())
	assertSucceed(t, examples.SagaFireRequest())
	assertSucceed(t, examples.SagaWaitFireRequest())
	assertSucceed(t, examples.TccBarrierFireRequest())
	assertSucceed(t, examples.TccFireRequest())
	assertSucceed(t, examples.TccFireRequestNested())
	assertSucceed(t, examples.XaFireRequest())
	assertSucceed(t, examples.MsgPbFireRequest())
}

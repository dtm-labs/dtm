package test

import (
	"testing"

	"github.com/yedf/dtm/examples"
)

func TestExamples(t *testing.T) {
	examples.QsStartSvr()
	for _, s := range examples.Samples {
		assertSucceed(t, s.Action())
	}
}

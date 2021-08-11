package test

import (
	"testing"

	"github.com/yedf/dtm/examples"
)

func TestExamples(t *testing.T) {
	// for coverage
	examples.QsStartSvr()
	for _, fn := range examples.Samples {
		assertSucceed(t, fn())
	}
}

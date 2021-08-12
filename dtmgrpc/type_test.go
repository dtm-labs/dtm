package dtmgrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	_, err := BarrierFromGrpc(&BusiRequest{Info: &BranchInfo{}})
	assert.Error(t, err)
}

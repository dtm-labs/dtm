package dtmgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	_, err := BarrierFromGrpc(context.Background())
	assert.Error(t, err)

	_, err = TccFromGrpc(context.Background())
	assert.Error(t, err)
}

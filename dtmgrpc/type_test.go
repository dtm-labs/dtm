package dtmgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
)

func TestType(t *testing.T) {
	_, err := BarrierFromGrpc(context.Background())
	assert.Error(t, err)

	_, err = TccFromGrpc(context.Background())
	assert.Error(t, err)

	old := GetCurrentDBType()
	SetCurrentDBType(dtmcli.DBTypeMysql)
	SetCurrentDBType(old)
}

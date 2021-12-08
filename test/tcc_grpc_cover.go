package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
)

func TestTccGrpcCoverNotConnected(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmgrpc.TccGlobalTransaction("localhost:01", gid, func(tcc *dtmgrpc.TccGrpc) error {
		return nil
	})
	assert.Error(t, err)
}

func TestTccGrpcCoverPanic(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmimp.CatchP(func() {
		_ = dtmgrpc.TccGlobalTransaction(examples.DtmHttpServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			panic("user panic")
		})
		assert.FailNow(t, "not executed")
	})
	assert.Contains(t, err.Error(), "user panic")
}

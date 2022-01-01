package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestTccCoverNotConnected(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction("localhost:01", gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		return nil, nil
	})
	assert.Error(t, err)
}

func TestTccCoverPanic(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmimp.CatchP(func() {
		_ = dtmcli.TccGlobalTransaction(dtmutil.DefaultHttpServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			panic("user panic")
		})
		assert.FailNow(t, "not executed")
	})
	assert.Contains(t, err.Error(), "user panic")
	waitTransProcessed(gid)
}

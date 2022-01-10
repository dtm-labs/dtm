package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
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
		_ = dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			panic("user panic")
		})
		assert.FailNow(t, "not executed")
	})
	assert.Contains(t, err.Error(), "user panic")
	waitTransProcessed(gid)
}

func TestTccNested(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TransInTccNested", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

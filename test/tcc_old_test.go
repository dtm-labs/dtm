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

func TestTccOldNormal(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOutOld", Busi+"/TransOutConfirmOld", Busi+"/TransOutRevertOld")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TransInOld", Busi+"/TransInConfirmOld", Busi+"/TransInRevertOld")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccOldRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, true)
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, rerr := tcc.CallBranch(req, Busi+"/TransOutOld", Busi+"/TransOutConfirmOld", Busi+"/TransOutRevertOld")
		assert.Nil(t, rerr)
		busi.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
		return tcc.CallBranch(req, Busi+"/TransInOld", Busi+"/TransInConfirmOld", Busi+"/TransInRevertOld")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusAborting, getTransStatus(gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccOldTimeout(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	timeoutChan := make(chan int, 1)

	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOutOld", Busi+"/TransOutConfirmOld", Busi+"/TransOutRevertOld")
		assert.Nil(t, err)
		go func() {
			cronTransOnceForwardNow(t, gid, 300)
			timeoutChan <- 0
		}()
		<-timeoutChan
		_, err = tcc.CallBranch(req, Busi+"/TransInOld", Busi+"/TransInConfirmOld", Busi+"/TransInRevertOld")
		assert.Error(t, err)
		return nil, err
	})
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

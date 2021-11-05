package test

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestXaLocalError(t *testing.T) {
	xc := examples.XaClient
	err := xc.XaGlobalTransaction("xaLocalError", func(xa *dtmcli.Xa) (*resty.Response, error) {
		return nil, fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
}

func TestXaNormal(t *testing.T) {
	xc := examples.XaClient
	gid := dtmimp.GetFuncName()
	err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := examples.GenTransReq(30, false, false)
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		if err != nil {
			return resp, err
		}
		return xa.CallBranch(req, examples.Busi+"/TransInXa")
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(gid))
}

func TestXaDuplicate(t *testing.T) {
	xc := examples.XaClient
	gid := dtmimp.GetFuncName()
	err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := examples.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		assert.Nil(t, err)
		sdb, err := dtmimp.StandaloneDB(common.DtmConfig.DB)
		assert.Nil(t, err)
		dtmimp.DBExec(sdb, "xa recover")
		dtmimp.DBExec(sdb, "xa commit 'xaDuplicate-0101'") // 先把某一个事务提交，模拟重复请求
		return xa.CallBranch(req, examples.Busi+"/TransInXa")
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(gid))
}
func TestXaRollback(t *testing.T) {
	xc := examples.XaClient
	gid := dtmimp.GetFuncName()
	err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := &examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure}
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		if err != nil {
			return resp, err
		}
		return xa.CallBranch(req, examples.Busi+"/TransInXa")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
}

func TestXaTimeout(t *testing.T) {
	xc := examples.XaClient
	gid := dtmimp.GetFuncName()
	timeoutChan := make(chan int, 1)
	err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		go func() {
			cronTransOnceForwardNow(1)
			cronTransOnceForwardNow(300)
			timeoutChan <- 0
		}()
		_ = <-timeoutChan
		return nil, nil
	})
	assert.Error(t, err)
}

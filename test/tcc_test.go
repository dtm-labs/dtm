/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

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

func TestTccNormal(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, true)
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, rerr := tcc.CallBranch(req, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, rerr)
		busi.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
		return tcc.CallBranch(req, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusAborting, getTransStatus(gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccTimeout(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	timeoutChan := make(chan int, 1)

	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, err)
		go func() {
			cronTransOnceForwardNow(t, gid, 300)
			timeoutChan <- 0
		}()
		<-timeoutChan
		_, err = tcc.CallBranch(req, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		assert.Error(t, err)
		return nil, err
	})
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccCompatible(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(dtmutil.DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))

}

func TestTccHeaders(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction2(dtmutil.DefaultHTTPServer, gid, func(t *dtmcli.Tcc) {
		t.BranchHeaders = map[string]string{
			"test_header": "test",
		}
	}, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		return tcc.CallBranch(req, Busi+"/TransOutHeaderYes", "", "")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

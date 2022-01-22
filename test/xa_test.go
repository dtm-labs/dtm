/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"fmt"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func getXc() *dtmcli.XaClient {
	return busi.XaClient
}

func TestXaNormal(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := busi.GenTransReq(30, false, false)
		resp, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		if err != nil {
			return resp, err
		}
		return xa.CallBranch(req, busi.Busi+"/TransInXa")
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestXaDuplicate(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := busi.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		assert.Nil(t, err)
		sdb, err := dtmimp.StandaloneDB(busi.BusiConf)
		assert.Nil(t, err)
		if dtmcli.GetCurrentDBType() == dtmcli.DBTypeMysql {
			_, err = dtmimp.DBExec(sdb, "xa recover")
			assert.Nil(t, err)
		}
		_, err = dtmimp.DBExec(sdb, dtmimp.GetDBSpecial().GetXaSQL("commit", gid+"-01")) // 先把某一个事务提交，模拟重复请求
		assert.Nil(t, err)
		return xa.CallBranch(req, busi.Busi+"/TransInXa")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestXaRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := busi.GenTransReq(30, false, true)
		resp, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		if err != nil {
			return resp, err
		}
		return xa.CallBranch(req, busi.Busi+"/TransInXa")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestXaLocalError(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		return nil, fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
	waitTransProcessed(gid)
}

func TestXaTimeout(t *testing.T) {
	gid := dtmimp.GetFuncName()
	timeoutChan := make(chan int, 1)
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		go func() {
			cronTransOnceForwardNow(t, gid, 300)
			timeoutChan <- 0
		}()
		<-timeoutChan
		return nil, nil
	})
	assert.Error(t, err)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{}, getBranchesStatus(gid))
}

func TestXaNotTimeout(t *testing.T) {
	gid := dtmimp.GetFuncName()
	timeoutChan := make(chan int, 1)
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		go func() {
			cronTransOnceForwardNow(t, gid, 0) // not timeout,
			timeoutChan <- 0
		}()
		<-timeoutChan
		req := busi.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		assert.Nil(t, err)
		busi.MainSwitch.NextResult.SetOnce(dtmcli.ResultOngoing) // make commit temp error
		return nil, nil
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

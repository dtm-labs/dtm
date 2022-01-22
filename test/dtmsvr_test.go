/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

var DtmServer = dtmutil.DefaultHTTPServer
var DtmGrpcServer = dtmutil.DefaultGrpcServer
var Busi = busi.Busi

func getTransStatus(gid string) string {
	return dtmsvr.GetTransGlobal(gid).Status
}

func getBranchesStatus(gid string) []string {
	branches := dtmsvr.GetStore().FindBranches(gid)
	status := []string{}
	for _, branch := range branches {
		status = append(status, branch.Status)
	}
	return status
}

func TestUpdateBranchAsync(t *testing.T) {
	if conf.Store.Driver != config.Mysql {
		return
	}
	conf.UpdateBranchSync = 0
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.WaitResult = true
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	time.Sleep(dtmsvr.UpdateBranchAsyncInterval)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	conf.UpdateBranchSync = 1
}

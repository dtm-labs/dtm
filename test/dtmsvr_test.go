/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"
	"time"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/examples"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var DtmServer = examples.DtmHttpServer
var Busi = examples.Busi
var app *gin.Engine

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

func assertSucceed(t *testing.T, gid string) {
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

func TestUpdateBranchAsync(t *testing.T) {
	if config.Store.Driver != "mysql" {
		return
	}
	common.Config.UpdateBranchSync = 0
	saga := genSaga1(dtmimp.GetFuncName(), false, false)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	time.Sleep(dtmsvr.UpdateBranchAsyncInterval)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	common.Config.UpdateBranchSync = 1
}

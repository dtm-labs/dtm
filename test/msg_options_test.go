/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgOptionsTimeout(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsg(gid)
	msg.Prepare("")
	cronTransOnce(t, gid)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(t, gid, 60)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutCustom(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsg(gid)
	msg.TimeoutToFail = 120
	msg.Prepare("")
	cronTransOnce(t, gid)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(t, gid, 60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutFailed(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsg(gid)
	msg.TimeoutToFail = 120
	msg.Prepare("")
	cronTransOnce(t, gid)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(t, gid, 60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

func TestMsgConcurrent(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Concurrent = true
	msg.Submit()
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

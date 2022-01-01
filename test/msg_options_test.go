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
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgOptionsTimeout(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	g := cronTransOnce()
	assert.Equal(t, msg.Gid, g)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutCustom(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.TimeoutToFail = 120
	msg.Prepare("")
	g := cronTransOnce()
	assert.Equal(t, msg.Gid, g)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutFailed(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.TimeoutToFail = 120
	msg.Prepare("")
	g := cronTransOnce()
	assert.Equal(t, msg.Gid, g)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

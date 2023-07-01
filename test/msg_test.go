/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"strings"
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgNormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Submit()
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgTimeoutSuccess(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsg(gid)
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgTimeoutFailed(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsg(gid)
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 360)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

func TestMsgAbnormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	err := msg.Prepare("")
	assert.Nil(t, err)
	err = msg.Submit()
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	err = msg.Prepare("")
	assert.Error(t, err)
}

func TestMsgTopicNotFoundFailed(t *testing.T) {
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultHTTPServer, dtmimp.GetFuncName()).
		AddTopic("non_existent_topic_TestMsgTopicNotFoundFailed", &req)
	msg.QueryPrepared = busi.Busi + "/QueryPrepared"
	assert.True(t, strings.Contains(msg.Submit().Error(), "topic not found"))
}

func genMsg(gid string) *dtmcli.Msg {
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultHTTPServer, gid).
		AddTopic("http_trans", &req)

	msg.QueryPrepared = busi.Busi + "/QueryPrepared"
	return msg
}

func subscribeTopic() {
	e2p(httpSubscribe("http_trans", busi.Busi+"/TransOut"))
	e2p(httpSubscribe("http_trans", busi.Busi+"/TransIn"))
}

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
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgGrpcNormal(t *testing.T) {
	msg := genGrpcMsg(dtmimp.GetFuncName())
	err := msg.Submit()
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
}

func TestMsgGrpcTimeoutSuccess(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genGrpcMsg(gid)
	err := msg.Prepare("")
	assert.Nil(t, err)
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
	cronTransOnce(t, gid)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
}

func TestMsgGrpcTimeoutFailed(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genGrpcMsg(gid)
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	busi.MainSwitch.QueryPreparedResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(t, gid, 180)
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
}

func genGrpcMsg(gid string) *dtmgrpc.MsgGrpc {
	req := &busi.BusiReq{Amount: 30}
	msg := dtmgrpc.NewMsgGrpc(dtmutil.DefaultGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransOut", req).
		Add(busi.BusiGrpc+"/busi.Busi/TransIn", req)
	msg.QueryPrepared = fmt.Sprintf("%s/busi.Busi/QueryPrepared", busi.BusiGrpc)
	return msg
}

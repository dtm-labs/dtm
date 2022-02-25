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
	"github.com/stretchr/testify/assert"
)

func TestMsgJrpcNormal(t *testing.T) {
	msg := genJrpcMsg(dtmimp.GetFuncName())
	msg.Submit()
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgJrpcRepeated(t *testing.T) {
	msg := genJrpcMsg(dtmimp.GetFuncName())
	msg.Submit()
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	err := msg.Submit()
	assert.Error(t, err)
}
func TestMsgJprcAbnormal(t *testing.T) {
	id := "no-use"
	resp, err := dtmcli.GetRestyClient().R().SetBody("hello").Post(dtmutil.DefaultJrpcServer)
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "-32700")

	resp, err = dtmcli.GetRestyClient().R().SetBody(map[string]string{
		"jsonrpc": "1.0",
		"method":  "newGid",
		"params":  "",
		"id":      id,
	}).Post(dtmutil.DefaultJrpcServer)
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "-32600")

	resp, err = dtmcli.GetRestyClient().R().SetBody(map[string]string{
		"jsonrpc": "2.0",
		"method":  "not-exists",
		"params":  "",
		"id":      id,
	}).Post(dtmutil.DefaultJrpcServer)
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "-32601")
}

func genJrpcMsg(gid string) *dtmcli.Msg {
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultJrpcServer, gid).
		Add(busi.Busi+"/TransOut", &req).
		Add(busi.Busi+"/TransIn", &req)
	msg.QueryPrepared = busi.Busi + "/QueryPrepared"
	msg.Protocol = "json-rpc"
	return msg
}

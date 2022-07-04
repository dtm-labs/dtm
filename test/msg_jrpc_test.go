/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"database/sql"
	"errors"
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

func TestMsgJrpcResults(t *testing.T) {
	msg := genJrpcMsg(dtmimp.GetFuncName())
	busi.MainSwitch.JrpcResult.SetOnce("OTHER")
	err := msg.Submit()
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	busi.MainSwitch.JrpcResult.SetOnce("ONGOING")
	cronTransOnceForwardNow(t, msg.Gid, 180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	busi.MainSwitch.JrpcResult.SetOnce("FAILURE")
	cronTransOnceForwardNow(t, msg.Gid, 180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))

	cronTransOnceForwardNow(t, msg.Gid, 180)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgJrpcDoAndSubmit(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultJrpcServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	msg.Protocol = dtmimp.Jrpc
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestMsgJrpcDoAndSubmitBusiFailed(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultJrpcServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	msg.Protocol = dtmimp.Jrpc
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
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

	_, err = dtmcli.GetRestyClient().R().SetBody("hello").Post("http://localhost:1001")
	assert.Error(t, err)

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

	resp, err = dtmcli.GetRestyClient().R().SetBody(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "registerBranch",
		"params": map[string]string{
			"trans_type": "not-exists",
		},
		"id": id,
	}).Post(dtmutil.DefaultJrpcServer)
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "-32603")
}

func TestMsgJprcAbnormal2(t *testing.T) {
	tb := dtmimp.NewTransBase(dtmimp.GetFuncName(), "msg", dtmutil.DefaultJrpcServer, "01")
	tb.Protocol = "json-rpc"
	_, err := dtmimp.TransCallDtmExt(tb, "", "newGid")
	assert.Nil(t, err)
}

func genJrpcMsg(gid string) *dtmcli.Msg {
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultJrpcServer, gid).
		Add(busi.Busi+"/TransOut", &req).
		Add(busi.BusiJrpcURL+"TransIn", &req)
	msg.QueryPrepared = busi.Busi + "/QueryPrepared"
	msg.Protocol = dtmimp.Jrpc
	return msg
}

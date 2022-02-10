package test

import (
	"errors"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgRedisDo(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err := msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "redis")
}

func TestMsgRedisDoBusiFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err := msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgRedisDoBusiLater(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	_, err := dtmcli.GetRestyClient().R().
		SetQueryParams(map[string]string{
			"trans_type": "msg",
			"gid":        gid,
			"branch_id":  "00",
			"op":         "msg",
			"barrier_id": "01",
		}).
		SetBody(req).Get(Busi + "/RedisQueryPrepared")
	assert.Nil(t, err)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err = msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
	})
	assert.Error(t, err, dtmcli.ErrDuplicated)
	assertSameBalance(t, before, "redis")
}

func TestMsgRedisDoPrepareFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer+"not-exists", gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err := msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgRedisDoCommitFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err := msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return errors.New("after commit error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgRedisDoCommitAfterFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaRedisTransIn", req)
	err := msg.DoAndSubmit(Busi+"/RedisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		err := bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
		dtmimp.E2P(err)
		return errors.New("an error")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "redis")
}

package test

import (
	"errors"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgGrpcRedisDo(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInRedis", req)
	err := msg.DoAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedRedis", func(bb *dtmcli.BranchBarrier) error {
		return bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "redis")
}

func TestMsgGrpcRedisDoBusiFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInRedis", req)
	err := msg.DoAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedRedis", func(bb *dtmcli.BranchBarrier) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgGrpcRedisDoPrepareFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer+"not-exists", gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInRedis", req)
	err := msg.DoAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedRedis", func(bb *dtmcli.BranchBarrier) error {
		return bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgGrpcRedisDoCommitFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInRedis", req)
	err := msg.DoAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedRedis", func(bb *dtmcli.BranchBarrier) error {
		return errors.New("after commit error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "redis")
}

func TestMsgGrpcRedisDoCommitAfterFailed(t *testing.T) {
	before := getBeforeBalances("redis")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInRedis", req)
	err := msg.DoAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedRedis", func(bb *dtmcli.BranchBarrier) error {
		err := bb.RedisCheckAdjustAmount(busi.RedisGet(), busi.GetRedisAccountKey(busi.TransOutUID), -30, 86400)
		dtmimp.E2P(err)
		return errors.New("an error")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "redis")
}

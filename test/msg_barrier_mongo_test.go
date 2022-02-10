package test

import (
	"errors"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMsgMongoDoSucceed(t *testing.T) {
	before := getBeforeBalances("mongo")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaMongoTransIn", req)
	err := msg.DoAndSubmit(Busi+"/MongoQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.MongoCall(busi.MongoGet(), func(sc mongo.SessionContext) error {
			return busi.SagaMongoAdjustBalance(sc, sc.Client(), busi.TransOutUID, -30, "")
		})
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "mongo")
}

func TestMsgMongoDoBusiFailed(t *testing.T) {
	before := getBeforeBalances("mongo")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaMongoTransIn", req)
	err := msg.DoAndSubmit(Busi+"/MongoQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mongo")
}

func TestMsgMongoDoBusiLater(t *testing.T) {
	before := getBeforeBalances("mongo")
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
		SetBody(req).Get(Busi + "/MongoQueryPrepared")
	assert.Nil(t, err)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaMongoTransIn", req)
	err = msg.DoAndSubmit(Busi+"/MongoQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.MongoCall(busi.MongoGet(), func(sc mongo.SessionContext) error {
			return busi.SagaMongoAdjustBalance(sc, sc.Client(), busi.TransOutUID, -30, "")
		})
	})
	assert.Error(t, err, dtmcli.ErrDuplicated)
	assertSameBalance(t, before, "mongo")
}

func TestMsgMongoDoCommitFailed(t *testing.T) {
	before := getBeforeBalances("mongo")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaMongoTransIn", req)
	err := msg.DoAndSubmit(Busi+"/MongoQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.MongoCall(busi.MongoGet(), func(sc mongo.SessionContext) error {
			err := busi.SagaMongoAdjustBalance(sc, sc.Client(), busi.TransOutUID, -30, "")
			assert.Nil(t, err)
			return errors.New("commit failed")
		})
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mongo")
}

func TestMsgMongoDoCommitAfterFailed(t *testing.T) {
	before := getBeforeBalances("mongo")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaMongoTransIn", req)
	err := msg.DoAndSubmit(Busi+"/MongoQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		err := bb.MongoCall(busi.MongoGet(), func(sc mongo.SessionContext) error {
			return busi.SagaMongoAdjustBalance(sc, sc.Client(), busi.TransOutUID, -30, "")
		})
		assert.Nil(t, err)
		return errors.New("an error")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "mongo")
}

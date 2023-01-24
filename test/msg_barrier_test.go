package test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/dtm-labs/logger"
	"github.com/stretchr/testify/assert"
)

func TestMsgDoAndSubmit(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitBusiFailed(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitBusiLater(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	_, err := dtmcli.GetRestyClient().R().
		SetQueryParams(map[string]string{
			"trans_type": "msg",
			"gid":        gid,
			"branch_id":  dtmimp.MsgDoBranch0,
			"op":         dtmimp.MsgDoOp,
			"barrier_id": dtmimp.MsgDoBarrier1,
		}).
		SetBody(req).Get(Busi + "/QueryPreparedB")
	assert.Nil(t, err)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err = msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return nil
	})
	assert.Error(t, err, dtmcli.ErrDuplicated)
	assertSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitPrepareFailed(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(DtmServer+"not-exists", gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitCommitFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var g *gomonkey.Patches
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		g = gomonkey.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			logger.Debugf("tx.Commit rollback and return error in test")
			_ = tx.Rollback()
			return errors.New("test error for patch")
		})
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	g.Reset()
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitCommitAfterFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenReqHTTP(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var guard *gomonkey.Patches
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		err := busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
		guard = gomonkey.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			guard.Reset()
			_ = tx.Commit()
			return errors.New("test error for patch")
		})
		return err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "mysql")
}

package test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgDoAndSubmit(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
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
	req := busi.GenTransReq(30, false, false)
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
	req := busi.GenTransReq(30, false, false)
	_, err := dtmcli.GetRestyClient().R().
		SetQueryParams(map[string]string{
			"trans_type": "msg",
			"gid":        gid,
			"branch_id":  "00",
			"op":         "msg",
			"barrier_id": "01",
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
	req := busi.GenTransReq(30, false, false)
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
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var g *monkey.PatchGuard
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		g = monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			logger.Debugf("tx.Commit rollback and return error in test")
			_ = tx.Rollback()
			return errors.New("test error for patch")
		})
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	g.Unpatch()
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgDoAndSubmitCommitAfterFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var guard *monkey.PatchGuard
	err := msg.DoAndSubmitDB(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		err := busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
		guard = monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			guard.Unpatch()
			_ = tx.Commit()
			return errors.New("test error for patch")
		})
		return err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "mysql")
}

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

func TestMsgPrepareAndSubmit(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestMsgPrepareAndSubmitBusiFailed(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return errors.New("an error")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgPrepareAndSubmitPrepareFailed(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer+"not-exists", gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

func TestMsgPrepareAndSubmitCommitFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var g *monkey.PatchGuard
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
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

func TestMsgPrepareAndSubmitCommitAfterFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var guard *monkey.PatchGuard
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		err := busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
		guard = monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			guard.Unpatch()
			_ = tx.Commit()
			return errors.New("test error for patch")
		})
		return err
	})
	assert.Nil(t, err) // final commit will ignore error after submit
	waitTransProcessed(gid)
	assertNotSameBalance(t, before, "mysql")
}

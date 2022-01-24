package test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgGrpcPrepareAndSubmit(t *testing.T) {
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInBSaga", req)
	err := msg.DoAndSubmitDB(busi.BusiGrpc+"/busi.Busi/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -int(req.Amount), "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before, "mysql")
}

func TestMsgGrpcPrepareAndSubmitCommitAfterFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInBSaga", req)
	var guard *monkey.PatchGuard
	err := msg.DoAndSubmitDB(busi.BusiGrpc+"/busi.Busi/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		err := busi.SagaAdjustBalance(tx, busi.TransOutUID, -int(req.Amount), "SUCCESS")
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

func TestMsgGrpcPrepareAndSubmitCommitFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances("mysql")
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.Busi+"/SagaBTransIn", req)
	var g *monkey.PatchGuard
	err := msg.DoAndSubmitDB(busi.BusiGrpc+"/busi.Busi/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		g = monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			logger.Debugf("tx.Commit rollback and return error in test")
			_ = tx.Rollback()
			return errors.New("test error for patch")
		})
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -int(req.Amount), "SUCCESS")
	})
	g.Unpatch()
	assert.Error(t, err)
	assertSameBalance(t, before, "mysql")
}

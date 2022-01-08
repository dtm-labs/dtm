package test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgGrpcPrepareAndSubmit(t *testing.T) {
	before := getBeforeBalances()
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInBSaga", req)
	err := msg.PrepareAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -int(req.Amount), "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before)
}

func TestMsgGrpcPrepareAndSubmitCommitAfterFailed(t *testing.T) {
	if conf.Store.IsDB() { // cannot patch tx.Commit, because Prepare also do Commit
		return
	}
	before := getBeforeBalances()
	gid := dtmimp.GetFuncName()
	req := busi.GenBusiReq(30, false, false)
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransInBSaga", req)
	var guard *monkey.PatchGuard
	err := msg.PrepareAndSubmit(busi.BusiGrpc+"/busi.Busi/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		err := busi.SagaAdjustBalance(tx, busi.TransOutUID, -int(req.Amount), "SUCCESS")
		guard = monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
			guard.Unpatch()
			_ = tx.Commit()
			return errors.New("test error for patch")
		})
		return err
	})
	assert.Error(t, err)
	cronTransOnceForwardNow(180)
	assertNotSameBalance(t, before)
}

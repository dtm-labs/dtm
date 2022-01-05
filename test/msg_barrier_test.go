package test

import (
	"database/sql"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestMsgPrepareAndSubmit(t *testing.T) {
	before := getBeforeBalances()
	gid := dtmcli.MustGenGid(DtmServer)
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(DtmServer, gid).
		Add(busi.Busi+"/SagaBTransIn1", req)
	err := msg.PrepareAndSubmit(Busi+"/QueryPreparedB", dbGet().ToSQLDB(), func(tx *sql.Tx) error {
		return busi.SagaAdjustBalance(tx, busi.TransOutUID, -req.Amount, "SUCCESS")
	})
	assert.Nil(t, err)
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
	assertNotSameBalance(t, before)
}

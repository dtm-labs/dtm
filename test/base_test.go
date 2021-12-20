/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

// BarrierModel barrier model for gorm
type BarrierModel struct {
	common.ModelBase
	dtmcli.BranchBarrier
}

// TableName gorm table name
func (BarrierModel) TableName() string { return "dtm_barrier.barrier" }

func TestBaseSqlDB(t *testing.T) {
	asserts := assert.New(t)
	db := common.DbGet(config.ExamplesDB)
	barrier := &dtmcli.BranchBarrier{
		TransType: "saga",
		Gid:       "gid2",
		BranchID:  "branch_id2",
		Op:        dtmcli.BranchAction,
	}
	db.Must().Exec("insert into dtm_barrier.barrier(trans_type, gid, branch_id, op, reason) values('saga', 'gid1', 'branch_id1', 'action', 'saga')")
	tx, err := db.ToSQLDB().Begin()
	asserts.Nil(err)
	err = barrier.Call(tx, func(tx *sql.Tx) error {
		dtmimp.Logf("rollback gid2")
		return fmt.Errorf("gid2 error")
	})
	asserts.Error(err, fmt.Errorf("gid2 error"))
	dbr := db.Model(&BarrierModel{}).Where("gid=?", "gid1").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
	dbr = db.Model(&BarrierModel{}).Where("gid=?", "gid2").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(0))
	barrier.BarrierID = 0
	err = barrier.CallWithDB(db.ToSQLDB(), func(tx *sql.Tx) error {
		dtmimp.Logf("submit gid2")
		return nil
	})
	asserts.Nil(err)
	dbr = db.Model(&BarrierModel{}).Where("gid=?", "gid2").Find(&[]BarrierModel{})
	asserts.Equal(dbr.RowsAffected, int64(1))
}

func TestBaseHttp(t *testing.T) {
	resp, err := dtmimp.RestyClient.R().SetQueryParam("panic_string", "1").Post(examples.Busi + "/TestPanic")
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "panic_string")
	resp, err = dtmimp.RestyClient.R().SetQueryParam("panic_error", "1").Post(examples.Busi + "/TestPanic")
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "panic_error")
}

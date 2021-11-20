/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Process process global transaction once
func (t *TransGlobal) Process(db *common.DB) map[string]interface{} {
	r := t.process(db)
	transactionMetrics(t, r["dtm_result"] == dtmcli.ResultSuccess)
	return r
}

func (t *TransGlobal) process(db *common.DB) map[string]interface{} {
	if t.Options != "" {
		dtmimp.MustUnmarshalString(t.Options, &t.TransOptions)
	}

	if !t.WaitResult {
		go t.processInner(db)
		return dtmcli.MapSuccess
	}
	submitting := t.Status == dtmcli.StatusSubmitted
	err := t.processInner(db)
	if err != nil {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": err.Error()}
	}
	if submitting && t.Status != dtmcli.StatusSucceed {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": "trans failed by user"}
	}
	return dtmcli.MapSuccess
}

func (t *TransGlobal) processInner(db *common.DB) (rerr error) {
	defer handlePanic(&rerr)
	defer func() {
		if rerr != nil {
			dtmimp.LogRedf("processInner got error: %s", rerr.Error())
		}
		if TransProcessedTestChan != nil {
			dtmimp.Logf("processed: %s", t.Gid)
			TransProcessedTestChan <- t.Gid
			dtmimp.Logf("notified: %s", t.Gid)
		}
	}()
	dtmimp.Logf("processing: %s status: %s", t.Gid, t.Status)
	branches := []TransBranch{}
	db.Must().Where("gid=?", t.Gid).Order("id asc").Find(&branches)
	t.lastTouched = time.Now()
	rerr = t.getProcessor().ProcessOnce(db, branches)
	return
}

func (t *TransGlobal) saveNew(db *common.DB) error {
	return db.Transaction(func(db1 *gorm.DB) error {
		db := &common.DB{DB: db1}
		t.setNextCron(cronReset)
		t.Options = dtmimp.MustMarshalString(t.TransOptions)
		if t.Options == "{}" {
			t.Options = ""
		}
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(t)
		if dbr.RowsAffected <= 0 { // 如果这个不是新事务，返回错误
			return errUniqueConflict
		}
		branches := t.getProcessor().GenBranches()
		if len(branches) > 0 {
			checkLocalhost(branches)
			db.Must().Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&branches)
		}
		return nil
	})
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"time"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
)

// Process process global transaction once
func (t *TransGlobal) Process() map[string]interface{} {
	r := t.process()
	transactionMetrics(t, r["dtm_result"] == dtmcli.ResultSuccess)
	return r
}

func (t *TransGlobal) process() map[string]interface{} {
	if t.Options != "" {
		dtmimp.MustUnmarshalString(t.Options, &t.TransOptions)
	}

	if !t.WaitResult {
		go t.processInner()
		return dtmcli.MapSuccess
	}
	submitting := t.Status == dtmcli.StatusSubmitted
	err := t.processInner()
	if err != nil {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": err.Error()}
	}
	if submitting && t.Status != dtmcli.StatusSucceed {
		return map[string]interface{}{"dtm_result": dtmcli.ResultFailure, "message": "trans failed by user"}
	}
	return dtmcli.MapSuccess
}

func (t *TransGlobal) processInner() (rerr error) {
	defer handlePanic(&rerr)
	defer func() {
		if rerr != nil && rerr != dtmcli.ErrOngoing {
			logger.Errorf("processInner got error: %s", rerr.Error())
		}
		if TransProcessedTestChan != nil {
			logger.Debugf("processed: %s", t.Gid)
			TransProcessedTestChan <- t.Gid
			logger.Debugf("notified: %s", t.Gid)
		}
	}()
	logger.Debugf("processing: %s status: %s", t.Gid, t.Status)
	branches := GetStore().FindBranches(t.Gid)
	t.lastTouched = time.Now()
	rerr = t.getProcessor().ProcessOnce(branches)
	return
}

func (t *TransGlobal) saveNew() error {
	branches := t.getProcessor().GenBranches()
	t.NextCronInterval = t.getNextCronInterval(cronReset)
	t.NextCronTime = common.GetNextTime(t.NextCronInterval)
	t.Options = dtmimp.MustMarshalString(t.TransOptions)
	if t.Options == "{}" {
		t.Options = ""
	}
	now := time.Now()
	t.CreateTime = &now
	t.UpdateTime = &now
	err := GetStore().MaySaveNewTrans(&t.TransGlobalStore, branches)
	logger.Infof("MaySaveNewTrans result: %v, global: %v branches: %v",
		err, t.TransGlobalStore.String(), dtmimp.MustMarshalString(branches))
	return err
}

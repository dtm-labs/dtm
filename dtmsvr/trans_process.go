/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/logger"
)

// Process process global transaction once
func (t *TransGlobal) Process(branches []TransBranch) error {
	r := t.process(branches)
	transactionMetrics(t, r == nil)
	return r
}

func (t *TransGlobal) process(branches []TransBranch) error {
	if t.Options != "" {
		dtmimp.MustUnmarshalString(t.Options, &t.TransOptions)
	}
	if t.ExtData != "" {
		dtmimp.MustUnmarshalString(t.ExtData, &t.Ext)
	}
	if !t.WaitResult {
		ctx := NewAsyncContext(t.Context)
		go func(ctx context.Context) {
			err := t.processInner(ctx, branches)
			if err != nil && !errors.Is(err, dtmimp.ErrOngoing) {
				logger.Errorf("processInner err: %v", err)
			}
		}(ctx)
		return nil
	}
	submitting := t.Status == dtmcli.StatusSubmitted
	err := t.processInner(t.Context, branches)
	if err != nil {
		return err
	}

	if submitting && t.Status != dtmcli.StatusSucceed {
		if t.RollbackReason != "" {
			return dtmcli.ErrorMessage2Error(t.RollbackReason, dtmcli.ErrFailure)
		}
		return fmt.Errorf("wait result not return success: %w", dtmcli.ErrFailure)
	}
	return nil
}

func (t *TransGlobal) processInner(ctx context.Context, branches []TransBranch) (rerr error) {
	defer handlePanic(&rerr)
	defer func() {
		if rerr != nil && !errors.Is(rerr, dtmcli.ErrOngoing) {
			logger.Errorf("processInner got error: %s", rerr.Error())
		}
		if TransProcessedTestChan != nil {
			logger.Debugf("processed: %s", t.Gid)
			TransProcessedTestChan <- t.Gid
			logger.Debugf("notified: %s", t.Gid)
		}
	}()
	logger.Debugf("processing: %s status: %s", t.Gid, t.Status)
	t.lastTouched = time.Now()
	rerr = t.getProcessor().ProcessOnce(ctx, branches)
	return
}

func (t *TransGlobal) saveNew() ([]TransBranch, error) {
	t.NextCronInterval = t.getNextCronInterval(cronReset)
	t.NextCronTime = dtmutil.GetNextTime(t.NextCronInterval)
	t.ExtData = dtmimp.MustMarshalString(t.Ext)
	if t.ExtData == "{}" {
		t.ExtData = ""
	}
	t.Options = dtmimp.MustMarshalString(t.TransOptions)
	if t.Options == "{}" {
		t.Options = ""
	}
	now := time.Now()
	t.CreateTime = &now
	t.UpdateTime = &now
	branches := t.getProcessor().GenBranches()
	for i := range branches {
		branches[i].CreateTime = &now
		branches[i].UpdateTime = &now
	}
	err := GetStore().MaySaveNewTrans(&t.TransGlobalStore, branches)
	logger.Infof("MaySaveNewTrans result: %v, global: %v branches: %v",
		err, t.TransGlobalStore.String(), dtmimp.MustMarshalString(branches))
	return branches, err
}

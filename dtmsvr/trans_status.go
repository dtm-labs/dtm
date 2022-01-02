/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtmdriver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (t *TransGlobal) touchCronTime(ctype cronType) {
	t.lastTouched = time.Now()
	GetStore().TouchCronTime(&t.TransGlobalStore, t.getNextCronInterval(ctype))
	logger.Infof("TouchCronTime for: %s", t.TransGlobalStore.String())
}

func (t *TransGlobal) changeStatus(status string) {
	updates := []string{"status", "update_time"}
	now := time.Now()
	if status == dtmcli.StatusSucceed {
		t.FinishTime = &now
		updates = append(updates, "finish_time")
	} else if status == dtmcli.StatusFailed {
		t.RollbackTime = &now
		updates = append(updates, "rollback_time")
	}
	t.UpdateTime = &now
	GetStore().ChangeGlobalStatus(&t.TransGlobalStore, status, updates, status == dtmcli.StatusSucceed || status == dtmcli.StatusFailed)
	logger.Infof("ChangeGlobalStatus to %s ok for %s", status, t.TransGlobalStore.String())
	t.Status = status
}

func (t *TransGlobal) changeBranchStatus(b *TransBranch, status string, branchPos int) {
	now := time.Now()
	b.Status = status
	b.FinishTime = &now
	b.UpdateTime = &now
	if conf.Store.Driver != dtmimp.DBTypeMysql && conf.Store.Driver != dtmimp.DBTypePostgres || conf.UpdateBranchSync > 0 || t.updateBranchSync {
		GetStore().LockGlobalSaveBranches(t.Gid, t.Status, []TransBranch{*b}, branchPos)
		logger.Infof("LockGlobalSaveBranches ok: gid: %s old status: %s branches: %s",
			b.Gid, dtmcli.StatusPrepared, b.String())
	} else { // 为了性能优化，把branch的status更新异步化
		updateBranchAsyncChan <- branchStatus{id: b.ID, gid: t.Gid, status: status, finishTime: &now}
	}
}

func (t *TransGlobal) isTimeout() bool {
	timeout := t.TimeoutToFail
	if t.TimeoutToFail == 0 && t.TransType != "saga" {
		timeout = conf.TimeoutToFail
	}
	if timeout == 0 {
		return false
	}
	return time.Since(*t.CreateTime)+NowForwardDuration >= time.Duration(timeout)*time.Second
}

func (t *TransGlobal) needProcess() bool {
	return t.Status == dtmcli.StatusSubmitted || t.Status == dtmcli.StatusAborting || t.Status == dtmcli.StatusPrepared && t.isTimeout()
}

func (t *TransGlobal) getURLResult(url string, branchID, op string, branchPayload []byte) (string, error) {
	if url == "" { // empty url is success
		return dtmcli.ResultSuccess, nil
	}
	if t.Protocol == "grpc" {
		dtmimp.PanicIf(strings.HasPrefix(url, "http"), fmt.Errorf("bad url for grpc: %s", url))
		server, method, err := dtmdriver.GetDriver().ParseServerMethod(url)
		if err != nil {
			return "", err
		}
		conn := dtmgimp.MustGetGrpcConn(server, true)
		ctx := dtmgimp.TransInfo2Ctx(t.Gid, t.TransType, branchID, op, "")
		kvs := dtmgimp.Map2Kvs(t.Ext.Headers)
		kvs = append(kvs, dtmgimp.Map2Kvs(t.BranchHeaders)...)
		ctx = metadata.AppendToOutgoingContext(ctx, kvs...)
		err = conn.Invoke(ctx, method, branchPayload, &[]byte{})
		if err == nil {
			return dtmcli.ResultSuccess, nil
		}
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Aborted {
			if st.Message() == dtmcli.ResultOngoing {
				return dtmcli.ResultOngoing, nil
			} else if st.Message() == dtmcli.ResultFailure {
				return dtmcli.ResultFailure, nil
			}
		}
		return "", err
	}
	dtmimp.PanicIf(!strings.HasPrefix(url, "http"), fmt.Errorf("bad url for http: %s", url))
	resp, err := dtmimp.RestyClient.R().SetBody(string(branchPayload)).
		SetQueryParams(map[string]string{
			"gid":        t.Gid,
			"trans_type": t.TransType,
			"branch_id":  branchID,
			"op":         op,
		}).
		SetHeader("Content-type", "application/json").
		SetHeaders(t.Ext.Headers).
		SetHeaders(t.TransOptions.BranchHeaders).
		Execute(dtmimp.If(branchPayload != nil || t.TransType == "xa", "POST", "GET").(string), url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (t *TransGlobal) getBranchResult(branch *TransBranch) (string, error) {
	body, err := t.getURLResult(branch.URL, branch.BranchID, branch.Op, branch.BinData)
	if err != nil {
		return "", err
	}
	if strings.Contains(body, dtmcli.ResultSuccess) {
		return dtmcli.StatusSucceed, nil
	} else if strings.HasSuffix(t.TransType, "saga") && branch.Op == dtmcli.BranchAction && strings.Contains(body, dtmcli.ResultFailure) {
		return dtmcli.StatusFailed, nil
	} else if strings.Contains(body, dtmcli.ResultOngoing) {
		return "", dtmimp.ErrOngoing
	}
	return "", fmt.Errorf("http result should contains SUCCESS|FAILURE|ONGOING. grpc error should return nil|Aborted with message(FAILURE|ONGOING). \nrefer to: https://dtm.pub/summary/arch.html#http\nunkown result will be retried: %s", body)
}

func (t *TransGlobal) execBranch(branch *TransBranch, branchPos int) error {
	status, err := t.getBranchResult(branch)
	if status != "" {
		t.changeBranchStatus(branch, status, branchPos)
	}
	branchMetrics(t, branch, status == dtmcli.StatusSucceed)
	// if time pass 1500ms and NextCronInterval is not default, then reset NextCronInterval
	if err == nil && time.Since(t.lastTouched)+NowForwardDuration >= 1500*time.Millisecond ||
		t.NextCronInterval > conf.RetryInterval && t.NextCronInterval > t.RetryInterval {
		t.touchCronTime(cronReset)
	} else if err == dtmimp.ErrOngoing {
		t.touchCronTime(cronKeep)
	} else if err != nil {
		t.touchCronTime(cronBackoff)
	}
	return err
}

func (t *TransGlobal) getNextCronInterval(ctype cronType) int64 {
	if ctype == cronBackoff {
		return t.NextCronInterval * 2
	} else if ctype == cronKeep {
		return t.NextCronInterval
	} else if t.RetryInterval != 0 {
		return t.RetryInterval
	} else {
		return conf.RetryInterval
	}
}

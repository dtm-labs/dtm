package dtmsvr

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtmdriver"
	"github.com/lithammer/shortuuid/v3"
	"google.golang.org/grpc/metadata"
)

// touchCronTime Based on ctype or delay set nextCronTime
// delay = 0 ,use ctype set nextCronTime and nextCronInterval
// delay > 0 ,use delay set nextCronTime ，use ctype set nextCronInterval
func (t *TransGlobal) touchCronTime(ctype cronType, delay uint64) {
	t.lastTouched = time.Now()
	nextCronInterval := t.getNextCronInterval(ctype)

	var nextCronTime *time.Time
	if delay > 0 {
		nextCronTime = dtmutil.GetNextTime(int64(delay))
	} else {
		nextCronTime = dtmutil.GetNextTime(nextCronInterval)
	}

	GetStore().TouchCronTime(&t.TransGlobalStore, nextCronInterval, nextCronTime)
	logger.Infof("TouchCronTime for: %s", t.TransGlobalStore.String())
}

type changeStatusParams struct {
	rollbackReason string
}
type changeStatusOption func(c *changeStatusParams)

func withRollbackReason(rollbackReason string) changeStatusOption {
	return func(c *changeStatusParams) {
		c.rollbackReason = rollbackReason
	}
}

func (t *TransGlobal) changeStatus(status string, opts ...changeStatusOption) {
	statusParams := &changeStatusParams{}
	for _, opt := range opts {
		opt(statusParams)
	}
	updates := []string{"status", "update_time"}
	now := time.Now()
	if status == dtmcli.StatusSucceed {
		t.FinishTime = &now
		updates = append(updates, "finish_time")
	} else if status == dtmcli.StatusFailed {
		t.RollbackTime = &now
		updates = append(updates, "rollback_time")
	}
	if statusParams.rollbackReason != "" {
		t.RollbackReason = statusParams.rollbackReason
		updates = append(updates, "rollback_reason")
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
	} else { // for better performance, batch the updates of branch status
		updateBranchAsyncChan <- branchStatus{id: b.ID, gid: t.Gid, branchID: b.BranchID, op: b.Op, status: status, finishTime: &now}
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

func (t *TransGlobal) needDelay(delay uint64) bool {
	return time.Since(*t.CreateTime)+CronForwardDuration < time.Duration(delay)*time.Second
}

func (t *TransGlobal) needProcess() bool {
	return t.Status == dtmcli.StatusSubmitted || t.Status == dtmcli.StatusAborting || t.Status == dtmcli.StatusPrepared && t.isTimeout()
}

func (t *TransGlobal) getURLResult(uri string, branchID, op string, branchPayload []byte) error {
	if uri == "" { // empty url is success
		return nil
	}
	if t.Protocol == dtmimp.ProtocolHTTP || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		if t.RequestTimeout != 0 {
			dtmimp.RestyClient.SetTimeout(time.Duration(t.RequestTimeout) * time.Second)
		}
		if t.Protocol == "json-rpc" && strings.Contains(uri, "method") {
			return t.getJSONRPCResult(uri, branchID, op, branchPayload)
		}
		return t.getHTTPResult(uri, branchID, op, branchPayload)
	}
	return t.getGrpcResult(uri, branchID, op, branchPayload)
}

func (t *TransGlobal) getHTTPResult(uri string, branchID, op string, branchPayload []byte) error {
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
		Execute(dtmimp.If(branchPayload != nil || t.TransType == "xa", "POST", "GET").(string), uri)
	if err != nil {
		return err
	}
	return dtmimp.RespAsErrorCompatible(resp)
}

func (t *TransGlobal) getJSONRPCResult(uri string, branchID, op string, branchPayload []byte) error {
	var params map[string]interface{}
	dtmimp.MustUnmarshal(branchPayload, &params)
	u, err := url.Parse(uri)
	dtmimp.E2P(err)
	params["gid"] = t.Gid
	params["trans_type"] = t.TransType
	params["branch_id"] = branchID
	params["op"] = op
	resp, err := dtmimp.RestyClient.R().SetBody(map[string]interface{}{
		"params":  params,
		"jsonrpc": "2.0",
		"method":  u.Query().Get("method"),
		"id":      shortuuid.New(),
	}).
		SetHeader("Content-type", "application/json").
		SetHeaders(t.Ext.Headers).
		SetHeaders(t.TransOptions.BranchHeaders).
		Post(uri)
	if err == nil {
		err = dtmimp.RespAsErrorCompatible(resp)
	}
	if err == nil {
		err = dtmimp.RespAsErrorByJSONRPC(resp)
	}
	return err
}

func (t *TransGlobal) getGrpcResult(uri string, branchID, op string, branchPayload []byte) error {
	// grpc handler
	server, method, err := dtmdriver.GetDriver().ParseServerMethod(uri)
	if err != nil {
		return err
	}

	conn := dtmgimp.MustGetGrpcConn(server, true)
	ctx := dtmgimp.TransInfo2Ctx(t.Context, t.Gid, t.TransType, branchID, op, "")
	kvs := dtmgimp.Map2Kvs(t.Ext.Headers)
	kvs = append(kvs, dtmgimp.Map2Kvs(t.BranchHeaders)...)
	ctx = metadata.AppendToOutgoingContext(ctx, kvs...)
	ctx = dtmgimp.RequestTimeoutNewContext(ctx, t.RequestTimeout)
	err = conn.Invoke(ctx, method, branchPayload, &[]byte{})
	if err == nil {
		return nil
	}
	return dtmgrpc.GrpcError2DtmError(err)
}

func (t *TransGlobal) getBranchResult(branch *TransBranch) (string, error) {
	err := t.getURLResult(branch.URL, branch.BranchID, branch.Op, branch.BinData)
	if err == nil {
		return dtmcli.StatusSucceed, nil
	} else if t.TransType == "saga" && branch.Op == dtmimp.OpAction && errors.Is(err, dtmcli.ErrFailure) {
		branch.Error = fmt.Errorf("url:%s return failed: %w", branch.URL, err)
		return dtmcli.StatusFailed, nil
	} else if errors.Is(err, dtmcli.ErrOngoing) {
		return "", dtmcli.ErrOngoing
	}
	return "", fmt.Errorf("http/grpc result should be specified as in:\nhttps://dtm.pub/summary/arch.html#http\nunkown result will be retried: %s", err)
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
		t.touchCronTime(cronReset, 0)
	} else if err == dtmimp.ErrOngoing {
		t.touchCronTime(cronKeep, 0)
	} else if err != nil {
		t.touchCronTime(cronBackoff, 0)
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
	} else if t.TimeoutToFail > 0 && t.TimeoutToFail < conf.RetryInterval {
		return t.TimeoutToFail
	} else {
		return conf.RetryInterval
	}
}

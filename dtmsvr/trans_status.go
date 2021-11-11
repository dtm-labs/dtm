package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (t *TransGlobal) touch(db *common.DB, ctype cronType) *gorm.DB {
	t.lastTouched = time.Now()
	updates := t.setNextCron(ctype)
	return db.Model(&TransGlobal{}).Where("gid=?", t.Gid).Select(updates).Updates(t)
}

func (t *TransGlobal) changeStatus(db *common.DB, status string) *gorm.DB {
	old := t.Status
	t.Status = status
	updates := t.setNextCron(cronReset)
	updates = append(updates, "status")
	now := time.Now()
	if status == dtmcli.StatusSucceed {
		t.FinishTime = &now
		updates = append(updates, "finish_time")
	} else if status == dtmcli.StatusFailed {
		t.RollbackTime = &now
		updates = append(updates, "rollback_time")
	}
	dbr := db.Must().Model(&TransGlobal{}).Where("status=? and gid=?", old, t.Gid).Select(updates).Updates(t)
	checkAffected(dbr)
	return dbr
}

func (t *TransGlobal) changeBranchStatus(db *common.DB, b *TransBranch, status string) {
	if common.DtmConfig.UpdateBranchSync > 0 || t.TransType == "saga" && t.TimeoutToFail > 0 {
		err := db.Transaction(func(tx *gorm.DB) error {
			dbr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&TransGlobal{}).Where("gid=? and status=?", t.Gid, t.Status).Find(&[]TransGlobal{})
			checkAffected(dbr) // check TransGlobal is not modified
			dbr = tx.Model(b).Updates(map[string]interface{}{
				"status":      status,
				"finish_time": time.Now(),
			})
			checkAffected(dbr)
			return dbr.Error
		})
		e2p(err)
	} else { // 为了性能优化，把branch的status更新异步化
		updateBranchAsyncChan <- branchStatus{id: b.ID, status: status}
	}
	b.Status = status
}

func (t *TransGlobal) isTimeout() bool {
	timeout := t.TimeoutToFail
	if t.TimeoutToFail == 0 && t.TransType != "saga" {
		timeout = config.TimeoutToFail
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
	if t.Protocol == "grpc" {
		dtmimp.PanicIf(strings.HasPrefix(url, "http"), fmt.Errorf("bad url for grpc: %s", url))
		server, method := dtmgimp.GetServerAndMethod(url)
		conn := dtmgimp.MustGetGrpcConn(server, true)
		ctx := dtmgimp.TransInfo2Ctx(t.Gid, t.TransType, branchID, op, "")
		err := conn.Invoke(ctx, method, branchPayload, []byte{})
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

func (t *TransGlobal) execBranch(db *common.DB, branch *TransBranch) error {
	status, err := t.getBranchResult(branch)
	if status != "" {
		t.changeBranchStatus(db, branch, status)
	}
	branchMetrics(t, branch, status == dtmcli.StatusSucceed)
	// if time pass 1500ms and NextCronInterval is not default, then reset NextCronInterval
	if err == nil && time.Since(t.lastTouched)+NowForwardDuration >= 1500*time.Millisecond ||
		t.NextCronInterval > config.RetryInterval && t.NextCronInterval > t.RetryInterval {
		t.touch(db, cronReset)
	} else if err == dtmimp.ErrOngoing {
		t.touch(db, cronKeep)
	} else if err != nil {
		t.touch(db, cronBackoff)
	}
	return err
}

func (t *TransGlobal) setNextCron(ctype cronType) []string {
	if ctype == cronBackoff {
		t.NextCronInterval = t.NextCronInterval * 2
	} else if ctype == cronKeep {
		// do nothing
	} else if t.RetryInterval != 0 {
		t.NextCronInterval = t.RetryInterval
	} else {
		t.NextCronInterval = config.RetryInterval
	}

	next := time.Now().Add(time.Duration(t.NextCronInterval) * time.Second)
	t.NextCronTime = &next
	return []string{"next_cron_interval", "next_cron_time"}
}

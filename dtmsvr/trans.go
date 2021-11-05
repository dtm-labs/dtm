package dtmsvr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errUniqueConflict = errors.New("unique key conflict error")

// TransGlobal global transaction
type TransGlobal struct {
	common.ModelBase
	Gid              string              `json:"gid"`
	TransType        string              `json:"trans_type"`
	Steps            []map[string]string `json:"steps" gorm:"-"`
	Payloads         []string            `json:"payloads" gorm:"-"`
	BinPayloads      [][]byte            `json:"-" gorm:"-"`
	Status           string              `json:"status"`
	QueryPrepared    string              `json:"query_prepared"`
	Protocol         string              `json:"protocol"`
	CommitTime       *time.Time
	FinishTime       *time.Time
	RollbackTime     *time.Time
	Options          string
	CustomData       string `json:"custom_data"`
	NextCronInterval int64
	NextCronTime     *time.Time
	dtmimp.TransOptions
	lastTouched time.Time // record the start time of process
}

// TableName TableName
func (*TransGlobal) TableName() string {
	return "dtm.trans_global"
}

type transProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(db *common.DB, branches []TransBranch) error
}

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
	dbr := db.Must().Model(t).Where("status=?", old).Select(updates).Updates(t)
	checkAffected(dbr)
	return dbr
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

// TransBranch branch transaction
type TransBranch struct {
	common.ModelBase
	Gid          string
	URL          string `json:"url"`
	BinData      []byte
	BranchID     string `json:"branch_id"`
	BranchType   string
	Status       string
	FinishTime   *time.Time
	RollbackTime *time.Time
}

// TableName TableName
func (*TransBranch) TableName() string {
	return "dtm.trans_branch"
}

func (t *TransBranch) changeStatus(db *common.DB, status string) *gorm.DB {
	if common.DtmConfig.UpdateBranchSync > 0 {
		dbr := db.Must().Model(t).Updates(map[string]interface{}{
			"status":      status,
			"finish_time": time.Now(),
		})
		checkAffected(dbr)
	} else { // 为了性能优化，把branch的status更新异步化
		updateBranchAsyncChan <- branchStatus{id: t.ID, status: status}
	}
	t.Status = status
	return db.DB
}

func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("rows affected 0, please check for abnormal trans"))
	}
}

type processorCreator func(*TransGlobal) transProcessor

var processorFac = map[string]processorCreator{}

func registorProcessorCreator(transType string, creator processorCreator) {
	processorFac[transType] = creator
}

func (t *TransGlobal) getProcessor() transProcessor {
	return processorFac[t.TransType](t)
}

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

type cronType int

const (
	cronBackoff cronType = iota
	cronReset
	cronKeep
)

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

func (t *TransGlobal) getURLResult(url string, branchID, branchType string, branchPayload []byte) (string, error) {
	if t.Protocol == "grpc" {
		dtmimp.PanicIf(strings.HasPrefix(url, "http"), fmt.Errorf("bad url for grpc: %s", url))
		server, method := dtmgimp.GetServerAndMethod(url)
		conn := dtmgimp.MustGetGrpcConn(server, true)
		ctx := dtmgimp.TransInfo2Ctx(t.Gid, t.TransType, branchID, branchType, "")
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
			"gid":         t.Gid,
			"trans_type":  t.TransType,
			"branch_id":   branchID,
			"branch_type": branchType,
		}).
		SetHeader("Content-type", "application/json").
		Execute(dtmimp.If(branchPayload != nil || t.TransType == "xa", "POST", "GET").(string), url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (t *TransGlobal) getBranchResult(branch *TransBranch) (string, error) {
	body, err := t.getURLResult(branch.URL, branch.BranchID, branch.BranchType, branch.BinData)
	if err != nil {
		return "", err
	}
	if strings.Contains(body, dtmcli.ResultSuccess) {
		return dtmcli.StatusSucceed, nil
	} else if strings.HasSuffix(t.TransType, "saga") && branch.BranchType == dtmcli.BranchAction && strings.Contains(body, dtmcli.ResultFailure) {
		return dtmcli.StatusFailed, nil
	} else if strings.Contains(body, dtmcli.ResultOngoing) {
		return "", dtmimp.ErrOngoing
	}
	return "", fmt.Errorf("http result should contains SUCCESS|FAILURE|ONGOING. grpc error should return nil|Aborted with message(FAILURE|ONGOING). \nrefer to: https://dtm.pub/summary/arch.html#http\nunkown result will be retried: %s", body)
}

func (t *TransGlobal) execBranch(db *common.DB, branch *TransBranch) error {
	status, err := t.getBranchResult(branch)
	if status != "" {
		branch.changeStatus(db, status)
	}
	branchMetrics(t, branch, status == dtmcli.StatusSucceed)
	// if time pass 1500ms and NextCronInterval is not default, then reset NextCronInterval
	if err == nil && time.Since(t.lastTouched)+NowForwardDuration >= 1500*time.Millisecond ||
		t.NextCronInterval > config.RetryInterval && t.NextCronInterval > t.RetryInterval {
		t.touch(db, cronReset)
	} else if err == dtmimp.ErrOngoing {
		t.touch(db, cronKeep)
	} else {
		t.touch(db, cronBackoff)
	}
	return err
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

// TransFromContext TransFromContext
func TransFromContext(c *gin.Context) *TransGlobal {
	b, err := c.GetRawData()
	e2p(err)
	m := TransGlobal{}
	dtmimp.MustUnmarshal(b, &m)
	dtmimp.Logf("creating trans in prepare")
	// Payloads will be store in BinPayloads, Payloads is only used to Unmarshal
	for _, p := range m.Payloads {
		m.BinPayloads = append(m.BinPayloads, []byte(p))
	}
	for _, d := range m.Steps {
		if d["data"] != "" {
			m.BinPayloads = append(m.BinPayloads, []byte(d["data"]))
		}
	}
	m.Protocol = "http"
	return &m
}

// TransFromDtmRequest TransFromContext
func TransFromDtmRequest(c *dtmgimp.DtmRequest) *TransGlobal {
	r := TransGlobal{
		Gid:           c.Gid,
		TransType:     c.TransType,
		QueryPrepared: c.QueryPrepared,
		Protocol:      "grpc",
		BinPayloads:   c.BinPayloads,
	}
	if c.Steps != "" {
		dtmimp.MustUnmarshalString(c.Steps, &r.Steps)
	}
	return &r
}

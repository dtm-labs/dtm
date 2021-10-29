package dtmsvr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errUniqueConflict = errors.New("unique key conflict error")

// TransGlobal global transaction
type TransGlobal struct {
	common.ModelBase
	Gid              string `json:"gid"`
	TransType        string `json:"trans_type"`
	Data             string `json:"data" gorm:"-"`
	Status           string `json:"status"`
	QueryPrepared    string `json:"query_prepared"`
	Protocol         string `json:"protocol"`
	CommitTime       *time.Time
	FinishTime       *time.Time
	RollbackTime     *time.Time
	Options          string
	CustomData       string `json:"custom_data"`
	NextCronInterval int64
	NextCronTime     *time.Time
	dtmcli.TransOptions
	processStarted time.Time // record the start time of process
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
	writeTransLog(t.Gid, "touch trans", "", "", "")
	updates := t.setNextCron(ctype)
	return db.Model(&TransGlobal{}).Where("gid=?", t.Gid).Select(updates).Updates(t)
}

func (t *TransGlobal) changeStatus(db *common.DB, status string) *gorm.DB {
	writeTransLog(t.Gid, "change status", status, "", "")
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

func (t *TransGlobal) getRetryInterval() int64 {
	if t.RetryInterval > 0 {
		return t.RetryInterval
	}
	return config.RetryInterval
}

func (t *TransGlobal) needProcess() bool {
	return t.Status == dtmcli.StatusSubmitted || t.Status == dtmcli.StatusAborting || t.Status == dtmcli.StatusPrepared && t.isTimeout()
}

// TransBranch branch transaction
type TransBranch struct {
	common.ModelBase
	Gid          string
	URL          string `json:"url"`
	Data         string
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
	writeTransLog(t.Gid, "branch change", status, t.BranchID, "")
	if common.DtmConfig.UpdateBranchSync > 0 {
		dbr := db.Must().Model(t).Updates(M{
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
		panic(fmt.Errorf("duplicate updating"))
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
func (t *TransGlobal) Process(db *common.DB) dtmcli.M {
	r := t.process(db)
	transactionMetrics(t, r["dtm_result"] == dtmcli.ResultSuccess)
	return r
}

func (t *TransGlobal) process(db *common.DB) dtmcli.M {
	if !t.WaitResult {
		go t.processInner(db)
		return dtmcli.MapSuccess
	}
	submitting := t.Status == dtmcli.StatusSubmitted
	err := t.processInner(db)
	if err != nil {
		return dtmcli.M{"dtm_result": dtmcli.ResultFailure, "message": err.Error()}
	}
	if submitting && t.Status != dtmcli.StatusSucceed {
		return dtmcli.M{"dtm_result": dtmcli.ResultFailure, "message": "trans failed by user"}
	}
	return dtmcli.MapSuccess
}

func (t *TransGlobal) processInner(db *common.DB) (rerr error) {
	defer handlePanic(&rerr)
	defer func() {
		if rerr != nil {
			dtmcli.LogRedf("processInner got error: %s", rerr.Error())
		}
		if TransProcessedTestChan != nil {
			dtmcli.Logf("processed: %s", t.Gid)
			TransProcessedTestChan <- t.Gid
			dtmcli.Logf("notified: %s", t.Gid)
		}
	}()
	dtmcli.Logf("processing: %s status: %s", t.Gid, t.Status)
	branches := []TransBranch{}
	db.Must().Where("gid=?", t.Gid).Order("id asc").Find(&branches)
	t.processStarted = time.Now()
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

func (t *TransGlobal) getURLResult(url string, branchID, branchType string, branchData []byte) (string, error) {
	if t.Protocol == "grpc" {
		dtmcli.PanicIf(strings.HasPrefix(url, "http"), fmt.Errorf("bad url for grpc: %s", url))
		server, method := dtmgrpc.GetServerAndMethod(url)
		conn := dtmgrpc.MustGetGrpcConn(server)
		err := conn.Invoke(context.Background(), method, &dtmgrpc.BusiRequest{
			Info: &dtmgrpc.BranchInfo{
				Gid:        t.Gid,
				TransType:  t.TransType,
				BranchID:   branchID,
				BranchType: branchType,
			},
			BusiData: branchData,
		}, &emptypb.Empty{})
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
	dtmcli.PanicIf(!strings.HasPrefix(url, "http"), fmt.Errorf("bad url for http: %s", url))
	resp, err := dtmcli.RestyClient.R().SetBody(string(branchData)).
		SetQueryParams(dtmcli.MS{
			"gid":         t.Gid,
			"trans_type":  t.TransType,
			"branch_id":   branchID,
			"branch_type": branchType,
		}).
		SetHeader("Content-type", "application/json").
		Execute(dtmcli.If(branchData == nil, "GET", "POST").(string), url)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (t *TransGlobal) getBranchResult(branch *TransBranch) (string, error) {
	body, err := t.getURLResult(branch.URL, branch.BranchID, branch.BranchType, []byte(branch.Data))
	if err != nil {
		return "", err
	}
	if strings.Contains(body, dtmcli.ResultSuccess) {
		return dtmcli.StatusSucceed, nil
	} else if strings.HasSuffix(t.TransType, "saga") && branch.BranchType == dtmcli.BranchAction && strings.Contains(body, dtmcli.ResultFailure) {
		return dtmcli.StatusFailed, nil
	} else if strings.Contains(body, dtmcli.ResultOngoing) {
		return "", dtmcli.ErrOngoing
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
	if err == nil && time.Since(t.processStarted)+NowForwardDuration >= 1500*time.Millisecond ||
		t.NextCronInterval > config.RetryInterval && t.NextCronInterval > t.RetryInterval {
		t.touch(db, cronReset)
	} else if err == dtmcli.ErrOngoing {
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
		writeTransLog(t.Gid, "create trans", t.Status, "", t.Data)
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(t)
		if dbr.RowsAffected <= 0 { // 如果这个不是新事务，返回错误
			return errUniqueConflict
		}
		branches := t.getProcessor().GenBranches()
		if len(branches) > 0 {
			writeTransLog(t.Gid, "save branches", t.Status, "", dtmcli.MustMarshalString(branches))
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
	data := M{}
	b, err := c.GetRawData()
	e2p(err)
	dtmcli.MustUnmarshal(b, &data)
	dtmcli.Logf("creating trans in prepare")
	if data["steps"] != nil {
		data["data"] = dtmcli.MustMarshalString(data["steps"])
	}
	m := TransGlobal{}
	dtmcli.MustRemarshal(data, &m)
	m.Options = dtmcli.MustMarshalString(m.TransOptions)
	if m.Options == "{}" {
		m.Options = ""
	}
	m.Protocol = "http"
	return &m
}

// TransFromDtmRequest TransFromContext
func TransFromDtmRequest(c *dtmgrpc.DtmRequest) *TransGlobal {
	return &TransGlobal{
		Gid:           c.Gid,
		TransType:     c.TransType,
		QueryPrepared: c.QueryPrepared,
		Data:          c.Data,
		Protocol:      "grpc",
	}
}

// TransFromDb construct trans from db
func TransFromDb(db *common.DB, gid string) *TransGlobal {
	m := TransGlobal{}
	dbr := db.Must().Model(&m).Where("gid=?", gid).First(&m)
	if dbr.Error == gorm.ErrRecordNotFound {
		return nil
	}
	e2p(dbr.Error)
	if m.Options != "" {
		dtmcli.MustUnmarshalString(m.Options, &m.TransOptions)
	}
	return &m
}

func checkLocalhost(branches []TransBranch) {
	if config.DisableLocalhost == 0 {
		return
	}
	for _, branch := range branches {
		if strings.HasPrefix(branch.URL, "http://localhost") || strings.HasPrefix(branch.URL, "localhost") {
			panic(errors.New("url for localhost is disabled. check for your config"))
		}
	}
}

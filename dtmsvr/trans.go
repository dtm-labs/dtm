package dtmsvr

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransGlobal struct {
	common.ModelBase
	Gid              string `json:"gid"`
	TransType        string `json:"trans_type"`
	Data             string `json:"data"`
	Status           string `json:"status"`
	QueryPrepared    string `json:"query_prepared"`
	CommitTime       *time.Time
	FinishTime       *time.Time
	RollbackTime     *time.Time
	NextCronInterval int64
	NextCronTime     *time.Time
}

func (*TransGlobal) TableName() string {
	return "dtm.trans_global"
}

type TransProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(db *common.DB, branches []TransBranch)
	ExecBranch(db *common.DB, branch *TransBranch)
}

func (t *TransGlobal) touch(db *common.DB, interval int64) *gorm.DB {
	writeTransLog(t.Gid, "touch trans", "", "", "")
	updates := t.setNextCron(interval)
	return db.Model(&TransGlobal{}).Where("gid=?", t.Gid).Select(updates).Updates(t)
}

func (t *TransGlobal) changeStatus(db *common.DB, status string) *gorm.DB {
	writeTransLog(t.Gid, "change status", status, "", "")
	old := t.Status
	t.Status = status
	updates := t.setNextCron(config.TransCronInterval)
	updates = append(updates, "status")
	now := time.Now()
	if status == "succeed" {
		t.FinishTime = &now
		updates = append(updates, "finish_time")
	} else if status == "failed" {
		t.RollbackTime = &now
		updates = append(updates, "rollback_time")
	}
	dbr := db.Must().Model(t).Where("status=?", old).Select(updates).Updates(t)
	checkAffected(dbr)
	return dbr
}

type TransBranch struct {
	common.ModelBase
	Gid          string
	Url          string
	Data         string
	BranchID     string `json:"branch_id"`
	BranchType   string
	Status       string
	FinishTime   *time.Time
	RollbackTime *time.Time
}

func (*TransBranch) TableName() string {
	return "dtm.trans_branch"
}

func (t *TransBranch) changeStatus(db *common.DB, status string) *gorm.DB {
	writeTransLog(t.Gid, "branch change", status, t.BranchID, "")
	dbr := db.Must().Model(t).Where("status=?", t.Status).Updates(M{
		"status":      status,
		"finish_time": time.Now(),
	})
	checkAffected(dbr)
	t.Status = status
	return dbr
}

func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("duplicate updating"))
	}
}

type processorCreator func(*TransGlobal) TransProcessor

var processorFac = map[string]processorCreator{}

func registorProcessorCreator(transType string, creator processorCreator) {
	processorFac[transType] = creator
}

func (trans *TransGlobal) getProcessor() TransProcessor {
	return processorFac[trans.TransType](trans)
}

func (trans *TransGlobal) Process(db *common.DB) {
	defer handlePanic()
	defer func() {
		if TransProcessedTestChan != nil {
			logrus.Printf("processed: %s", trans.Gid)
			TransProcessedTestChan <- trans.Gid
		}
	}()
	logrus.Printf("processing: %s", trans.Gid)
	branches := []TransBranch{}
	db.Must().Where("gid=?", trans.Gid).Order("id asc").Find(&branches)
	trans.getProcessor().ProcessOnce(db, branches)
}

func (trans *TransGlobal) getBranchParams(branch *TransBranch) common.MS {
	return common.MS{
		"gid":         trans.Gid,
		"trans_type":  trans.TransType,
		"branch_id":   branch.BranchID,
		"branch_type": branch.BranchType,
	}
}

func (t *TransGlobal) setNextCron(expireIn int64) []string {
	t.NextCronInterval = expireIn
	next := time.Now().Add(time.Duration(config.TransCronInterval) * time.Second)
	t.NextCronTime = &next
	return []string{"next_cron_interval", "next_cron_time"}
}

func (t *TransGlobal) SaveNew(db *common.DB) {
	if t.Gid == "" {
		t.Gid = GenGid()
	}
	err := db.Transaction(func(db1 *gorm.DB) error {
		db := &common.DB{DB: db1}
		updates := t.setNextCron(config.TransCronInterval)
		writeTransLog(t.Gid, "create trans", t.Status, "", t.Data)
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(t)
		if dbr.RowsAffected > 0 { // 如果这个是新事务，保存所有的分支
			branches := t.getProcessor().GenBranches()
			if len(branches) > 0 {
				writeTransLog(t.Gid, "save branches", t.Status, "", common.MustMarshalString(branches))
				db.Must().Clauses(clause.OnConflict{
					DoNothing: true,
				}).Create(&branches)
			}
		} else if dbr.RowsAffected == 0 && t.Status == "submitted" { // 如果数据库已经存放了prepared的事务，则修改状态
			dbr = db.Must().Model(t).Where("gid=? and status=?", t.Gid, "prepared").Select(append(updates, "status")).Updates(t)
		}
		return nil
	})
	e2p(err)
}

func TransFromContext(c *gin.Context) *TransGlobal {
	data := M{}
	b, err := c.GetRawData()
	e2p(err)
	common.MustUnmarshal(b, &data)
	logrus.Printf("creating trans in prepare")
	if data["steps"] != nil {
		data["data"] = common.MustMarshalString(data["steps"])
	}
	m := TransGlobal{}
	common.MustRemarshal(data, &m)
	return &m
}

func TransFromDb(db *common.DB, gid string) *TransGlobal {
	m := TransGlobal{}
	dbr := db.Must().Model(&m).Where("gid=?", gid).First(&m)
	e2p(dbr.Error)
	return &m
}

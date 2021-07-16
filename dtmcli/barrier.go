package dtmcli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
)

// BusiFunc type for busi func
type BusiFunc func(db *sql.DB) (interface{}, error)

// TransInfo every branch info
type TransInfo struct {
	TransType  string
	Gid        string
	BranchID   string
	BranchType string
}

func (t *TransInfo) String() string {
	return fmt.Sprintf("transInfo: %s %s %s %s", t.TransType, t.Gid, t.BranchID, t.BranchType)
}

// TransInfoFromReq construct transaction info from request
func TransInfoFromReq(c *gin.Context) *TransInfo {
	ti := &TransInfo{
		TransType:  c.Query("trans_type"),
		Gid:        c.Query("gid"),
		BranchID:   c.Query("branch_id"),
		BranchType: c.Query("branch_type"),
	}
	if ti.TransType == "" || ti.Gid == "" || ti.BranchID == "" || ti.BranchType == "" {
		panic(fmt.Errorf("invlid trans info: %v", ti))
	}
	return ti
}

// BarrierModel barrier model for gorm
type BarrierModel struct {
	common.ModelBase
	TransInfo
}

// TableName gorm table name
func (BarrierModel) TableName() string { return "dtm_barrier.barrier" }

func insertBarrier(tx *sql.Tx, transType string, gid string, branchID string, branchType string) (int64, error) {
	if branchType == "" {
		return 0, nil
	}
	res, err := tx.Exec("insert into dtm_barrier.barrier(trans_type, gid, branch_id, branch_type) values(?,?,?,?)", transType, gid, branchID, branchType)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ThroughBarrierCall barrier interface. busiCall will be called only when the request is necessary
func ThroughBarrierCall(db *sql.DB, transInfo *TransInfo, busiCall BusiFunc) (res interface{}, rerr error) {
	tx, rerr := db.BeginTx(context.Background(), &sql.TxOptions{})
	if rerr != nil {
		return
	}
	defer func() {
		if x := recover(); x != nil {
			tx.Rollback()
			panic(x)
		} else if rerr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	originType := map[string]string{
		"cancel":     "action",
		"compensate": "action",
	}[transInfo.BranchType]
	originAffected, _ := insertBarrier(tx, transInfo.TransType, transInfo.Gid, transInfo.BranchID, originType)
	currentAffected, rerr := insertBarrier(tx, transInfo.TransType, transInfo.Gid, transInfo.BranchID, transInfo.BranchType)
	if currentAffected == 0 || (originType == "cancel" || originType == "compensate") && originAffected > 0 {
		res = "SUCCESS" // 如果被忽略，那么直接返回 "SUCCESS"，表示成功，可以进行下一步
		return
	}
	res, rerr = busiCall(db)
	return
}

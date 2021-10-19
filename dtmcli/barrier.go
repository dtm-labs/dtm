package dtmcli

import (
	"fmt"
	"net/url"
)

// BusiFunc type for busi func
type BusiFunc func(db DB) error

// BranchBarrier every branch info
type BranchBarrier struct {
	TransType  string
	Gid        string
	BranchID   string
	BranchType string
	BarrierID  int
}

func (bb *BranchBarrier) String() string {
	return fmt.Sprintf("transInfo: %s %s %s %s", bb.TransType, bb.Gid, bb.BranchID, bb.BranchType)
}

// BarrierFromQuery construct transaction info from request
func BarrierFromQuery(qs url.Values) (*BranchBarrier, error) {
	return BarrierFrom(qs.Get("trans_type"), qs.Get("gid"), qs.Get("branch_id"), qs.Get("branch_type"))
}

// BarrierFrom construct transaction info from request
func BarrierFrom(transType, gid, branchID, branchType string) (*BranchBarrier, error) {
	ti := &BranchBarrier{
		TransType:  transType,
		Gid:        gid,
		BranchID:   branchID,
		BranchType: branchType,
	}
	if ti.TransType == "" || ti.Gid == "" || ti.BranchID == "" || ti.BranchType == "" {
		return nil, fmt.Errorf("invlid trans info: %v", ti)
	}
	return ti, nil
}

func insertBarrier(tx Tx, transType string, gid string, branchID string, branchType string, barrierID string, reason string) (int64, error) {
	if branchType == "" {
		return 0, nil
	}
	sql := GetDBSpecial().GetInsertIgnoreTemplate("dtm_barrier.barrier(trans_type, gid, branch_id, branch_type, barrier_id, reason) values(?,?,?,?,?,?)", "uniq_barrier")
	return DBExec(tx, sql, transType, gid, branchID, branchType, barrierID, reason)
}

// Call 子事务屏障，详细介绍见 https://zhuanlan.zhihu.com/p/388444465
// tx: 本地数据库的事务对象，允许子事务屏障进行事务操作
// bisiCall: 业务函数，仅在必要时被调用
func (bb *BranchBarrier) Call(tx Tx, busiCall BusiFunc) (rerr error) {
	bb.BarrierID = bb.BarrierID + 1
	bid := fmt.Sprintf("%02d", bb.BarrierID)
	defer func() {
		// Logf("barrier call error is %v", rerr)
		if x := recover(); x != nil {
			tx.Rollback()
			panic(x)
		} else if rerr != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	ti := bb
	originType := map[string]string{
		BranchCancel:     BranchTry,
		BranchCompensate: BranchAction,
	}[ti.BranchType]
	originAffected, _ := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, originType, bid, ti.BranchType)
	currentAffected, rerr := insertBarrier(tx, ti.TransType, ti.Gid, ti.BranchID, ti.BranchType, bid, ti.BranchType)
	Logf("originAffected: %d currentAffected: %d", originAffected, currentAffected)
	if (ti.BranchType == BranchCancel || ti.BranchType == BranchCompensate) && originAffected > 0 || // 这个是空补偿
		currentAffected == 0 { // 这个是重复请求或者悬挂
		return
	}
	rerr = busiCall(tx)
	return
}

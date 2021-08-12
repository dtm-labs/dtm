package dtmgrpc

import (
	"database/sql"

	"github.com/yedf/dtm/dtmcli"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// BranchBarrier 子事务屏障
type BranchBarrier struct {
	*dtmcli.BranchBarrier
}

// Call 子事务屏障，详细介绍见 https://zhuanlan.zhihu.com/p/388444465
// db: 本地数据库
// transInfo: 事务信息
// bisiCall: 业务函数，仅在必要时被调用
// 返回值:
// 如果发生悬挂，则busiCall不会被调用，直接返回错误 ErrFailure，全局事务尽早进行回滚
// 如果正常调用，重复调用，空补偿，返回的错误值为nil，正常往下进行
func (bb *BranchBarrier) Call(db *sql.DB, busiCall dtmcli.BusiFunc) (rerr error) {
	err := bb.BranchBarrier.Call(db, busiCall)
	if err == dtmcli.ErrFailure {
		return status.New(codes.Aborted, "user rollback").Err()
	}
	return err
}

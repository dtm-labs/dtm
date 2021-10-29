package dtmgrpc

import (
	"github.com/yedf/dtm/dtmcli"
)

// BranchBarrier 子事务屏障
type BranchBarrier struct {
	*dtmcli.BranchBarrier
}

// BarrierFromGrpc 从BusiRequest生成一个Barrier
func BarrierFromGrpc(in *BusiRequest) (*BranchBarrier, error) {
	b, err := dtmcli.BarrierFrom(in.Info.TransType, in.Info.Gid, in.Info.BranchID, in.Info.BranchType)
	return &BranchBarrier{
		BranchBarrier: b,
	}, err
}

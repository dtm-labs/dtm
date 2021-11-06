package dtmgrpc

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
)

// BarrierFromGrpc generate a Barrier from grpc context
func BarrierFromGrpc(ctx context.Context) (*dtmcli.BranchBarrier, error) {
	tb := dtmgimp.TransBaseFromGrpc(ctx)
	return dtmcli.BarrierFrom(tb.TransType, tb.Gid, tb.BranchID, tb.Op)
}

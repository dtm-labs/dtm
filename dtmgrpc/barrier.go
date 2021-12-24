/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"context"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
)

// BarrierFromGrpc generate a Barrier from grpc context
func BarrierFromGrpc(ctx context.Context) (*dtmcli.BranchBarrier, error) {
	tb := dtmgimp.TransBaseFromGrpc(ctx)
	return dtmcli.BarrierFrom(tb.TransType, tb.Gid, tb.BranchID, tb.Op)
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/dtmcli/logger"
	dtmgrpc "github.com/dtm-labs/dtm/dtmgrpc"
)

func init() {
	addSample("grpc_saga", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSagaGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
	addSample("grpc_saga_wait", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSagaGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		saga.WaitResult = true
		err := saga.Submit()
		logger.FatalIfError(err)
		return saga.Gid
	})
}

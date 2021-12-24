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
	addSample("grpc_msg", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", req)
		err := msg.Submit()
		logger.FatalIfError(err)
		return msg.Gid
	})
}

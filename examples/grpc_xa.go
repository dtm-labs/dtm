/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	context "context"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	addSample("grpc_xa", func() string {
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		req := &BusiReq{Amount: 30}
		err := XaGrpcClient.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
			r := &emptypb.Empty{}
			err := xa.CallBranch(req, BusiGrpc+"/examples.Busi/TransOutXa", r)
			if err != nil {
				return err
			}
			err = xa.CallBranch(req, BusiGrpc+"/examples.Busi/TransInXa", r)
			return err
		})
		logger.FatalIfError(err)
		return gid
	})
}

func (s *busiServer) XaNotify(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return XaGrpcClient.HandleCallback(ctx)
}

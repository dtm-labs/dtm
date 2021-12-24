/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/dtmcli/logger"
	dtmgrpc "github.com/dtm-labs/dtm/dtmgrpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	addSample("grpc_tcc", func() string {
		logger.Debugf("tcc simple transaction begin")
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		err := dtmgrpc.TccGlobalTransaction(DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			data := &BusiReq{Amount: 30}
			r := &emptypb.Empty{}
			err := tcc.CallBranch(data, BusiGrpc+"/examples.Busi/TransOutTcc", BusiGrpc+"/examples.Busi/TransOutConfirm", BusiGrpc+"/examples.Busi/TransOutRevert", r)
			if err != nil {
				return err
			}
			err = tcc.CallBranch(data, BusiGrpc+"/examples.Busi/TransInTcc", BusiGrpc+"/examples.Busi/TransInConfirm", BusiGrpc+"/examples.Busi/TransInRevert", r)
			return err
		})
		logger.FatalIfError(err)
		return gid
	})
}

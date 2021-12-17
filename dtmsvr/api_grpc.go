/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	pb "github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/protobuf/types/known/emptypb"
)

// dtmServer is used to implement dtmgimp.DtmServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

func (s *dtmServer) NewGid(ctx context.Context, in *emptypb.Empty) (*dtmgimp.DtmGidReply, error) {
	return &dtmgimp.DtmGidReply{Gid: GenGid()}, nil
}

func (s *dtmServer) Submit(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcSubmit(TransFromDtmRequest(in))
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

func (s *dtmServer) Prepare(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcPrepare(TransFromDtmRequest(in))
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

func (s *dtmServer) Abort(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcAbort(TransFromDtmRequest(in))
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

func (s *dtmServer) RegisterBranch(ctx context.Context, in *pb.DtmBranchRequest) (*emptypb.Empty, error) {
	r, err := svcRegisterBranch(in.TransType, &TransBranch{
		Gid:      in.Gid,
		BranchID: in.BranchID,
		Status:   dtmcli.StatusPrepared,
		BinData:  in.BusiPayload,
	}, in.Data)
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

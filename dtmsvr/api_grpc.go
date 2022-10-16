/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmgrpc"
	pb "github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// dtmServer is used to implement dtmgimp.DtmServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

func (s *dtmServer) NewGid(ctx context.Context, in *emptypb.Empty) (*pb.DtmGidReply, error) {
	return &pb.DtmGidReply{Gid: GenGid()}, nil
}

func (s *dtmServer) Submit(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r := svcSubmit(TransFromDtmRequest(ctx, in))
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(r)
}

func (s *dtmServer) Prepare(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r := svcPrepare(TransFromDtmRequest(ctx, in))
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(r)
}

func (s *dtmServer) Abort(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r := svcAbort(TransFromDtmRequest(ctx, in))
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(r)
}

func (s *dtmServer) RegisterBranch(ctx context.Context, in *pb.DtmBranchRequest) (*emptypb.Empty, error) {
	r := svcRegisterBranch(in.TransType, &TransBranch{
		Gid:      in.Gid,
		BranchID: in.BranchID,
		Status:   dtmcli.StatusPrepared,
		BinData:  in.BusiPayload,
	}, in.Data)
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(r)
}

func (s *dtmServer) PrepareWorkflow(ctx context.Context, in *pb.DtmRequest) (*pb.DtmProgressesReply, error) {
	trans, branches, err := svcPrepareWorkflow(TransFromDtmRequest(ctx, in))
	reply := &pb.DtmProgressesReply{
		Transaction: &pb.DtmTransaction{
			Gid:            trans.Gid,
			Status:         trans.Status,
			RollbackReason: trans.RollbackReason,
		},
		Progresses: []*pb.DtmProgress{},
	}
	for _, b := range branches {
		reply.Progresses = append(reply.Progresses, &pb.DtmProgress{
			Status:   b.Status,
			BranchID: b.BranchID,
			Op:       b.Op,
			BinData:  b.BinData,
		})
	}
	return reply, dtmgrpc.DtmError2GrpcError(err)
}

func (s *dtmServer) Subscribe(ctx context.Context, in *pb.DtmTopicRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(Subscribe(in.Topic, in.URL, in.Remark))
}

func (s *dtmServer) Unsubscribe(ctx context.Context, in *pb.DtmTopicRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(Unsubscribe(in.Topic, in.URL))
}

func (s *dtmServer) DeleteTopic(ctx context.Context, in *pb.DtmTopicRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(GetStore().DeleteKV(topicsCat, in.Topic))
}

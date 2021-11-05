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

func (s *dtmServer) RegisterTccBranch(ctx context.Context, in *pb.DtmTccBranchRequest) (*emptypb.Empty, error) {
	r, err := svcRegisterTccBranch(&TransBranch{
		Gid:      in.Info.Gid,
		BranchID: in.Info.BranchID,
		Status:   dtmcli.StatusPrepared,
		BinData:  in.BusiPayload,
	}, map[string]string{
		dtmcli.BranchCancel:  in.Cancel,
		dtmcli.BranchConfirm: in.Confirm,
		dtmcli.BranchTry:     in.Try,
	})
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

func (s *dtmServer) RegisterXaBranch(ctx context.Context, in *pb.DtmXaBranchRequest) (*emptypb.Empty, error) {
	r, err := svcRegisterXaBranch(&TransBranch{
		Gid:      in.Info.Gid,
		BranchID: in.Info.BranchID,
		Status:   dtmcli.StatusPrepared,
		BinData:  in.BusiPayload,
		URL:      in.Notify,
	})
	return &emptypb.Empty{}, dtmgimp.Result2Error(r, err)
}

package dtmsvr

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	pb "github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// dtmServer is used to implement helloworld.GreeterServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

func (s *dtmServer) NewGid(ctx context.Context, in *emptypb.Empty) (*dtmgrpc.DtmGidReply, error) {
	return &dtmgrpc.DtmGidReply{Gid: GenGid()}, nil
}

func (s *dtmServer) Submit(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcSubmit(TransFromDtmRequest(in), in.WaitResult)
	return &emptypb.Empty{}, dtmgrpc.Result2Error(r, err)
}

func (s *dtmServer) Prepare(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcPrepare(TransFromDtmRequest(in))
	return &emptypb.Empty{}, dtmgrpc.Result2Error(r, err)
}

func (s *dtmServer) Abort(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	r, err := svcAbort(TransFromDtmRequest(in), in.WaitResult)
	return &emptypb.Empty{}, dtmgrpc.Result2Error(r, err)
}

func (s *dtmServer) RegisterTccBranch(ctx context.Context, in *pb.DtmTccBranchRequest) (*emptypb.Empty, error) {
	r, err := svcRegisterTccBranch(&TransBranch{
		Gid:      in.Info.Gid,
		BranchID: in.Info.BranchID,
		Status:   "prepared",
		Data:     in.BusiData,
	}, dtmcli.MS{
		"cancel":  in.Cancel,
		"confirm": in.Confirm,
		"try":     in.Try,
	})
	return &emptypb.Empty{}, dtmgrpc.Result2Error(r, err)
}

func (s *dtmServer) RegisterXaBranch(ctx context.Context, in *pb.DtmXaBranchRequest) (*emptypb.Empty, error) {
	r, err := svcRegisterXaBranch(&TransBranch{
		Gid:      in.Info.Gid,
		BranchID: in.Info.BranchID,
		Status:   "prepared",
		Data:     in.BusiData,
		URL:      in.Notify,
	})
	return &emptypb.Empty{}, dtmgrpc.Result2Error(r, err)
}

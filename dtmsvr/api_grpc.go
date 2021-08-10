package dtmsvr

import (
	"context"

	pb "github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// dtmServer is used to implement helloworld.GreeterServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

func (s *dtmServer) Submit(ctx context.Context, in *pb.DtmRequest) (*emptypb.Empty, error) {
	svcSubmit(TransFromDtmRequest(in), in.WaitResult)
	return &emptypb.Empty{}, nil
}

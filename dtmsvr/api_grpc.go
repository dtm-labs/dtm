package dtmsvr

import (
	"context"
	"log"

	pb "github.com/yedf/dtm/dtmpb"
)

// dtmServer is used to implement helloworld.GreeterServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

func (s *dtmServer) Call(ctx context.Context, in *pb.DtmRequest) (*pb.DtmReply, error) {
	log.Printf("dtmServer Received: %v", in)
	return &pb.DtmReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

func (s *dtmServer) Submit(ctx context.Context, in *pb.DtmRequest) (*pb.DtmReply, error) {
	svcSubmit(TransFromDtmRequest(in), in.WaitResult)
	return &pb.DtmReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

package dtmsvr

import (
	"context"
	"log"

	pb "github.com/yedf/dtm/dtmcli"
)

// dtmServer is used to implement helloworld.GreeterServer.
type dtmServer struct {
	pb.UnimplementedDtmServer
}

// SayHello implements helloworld.GreeterServer
func (s *dtmServer) Call(ctx context.Context, in *pb.DtmRequest) (*pb.DtmReply, error) {
	log.Printf("Received: %v", in)
	return &pb.DtmReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

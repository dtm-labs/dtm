package dtmsvr

import (
	"context"
	"log"

	pb "github.com/yedf/dtm/dtmcli"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedDtmServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Call(ctx context.Context, in *pb.DtmRequest) (*pb.DtmReply, error) {
	log.Printf("Received: %v", in)
	return &pb.DtmReply{}, nil
}

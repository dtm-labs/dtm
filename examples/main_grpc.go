package examples

import (
	"context"
	"log"

	dtmpb "github.com/yedf/dtm/dtmpb"
)

// busiServer is used to implement helloworld.GreeterServer.
type busiServer struct {
	UnimplementedBusiServer
}

// SayHello implements helloworld.GreeterServer
func (s *busiServer) Call(ctx context.Context, in *dtmpb.BusiRequest) (*dtmpb.BusiReply, error) {
	log.Printf("busiServer received: %v", in)
	return &dtmpb.BusiReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

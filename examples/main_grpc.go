package examples

import (
	"context"
	"log"

	dtmcli "github.com/yedf/dtm/dtmcli"
)

// busiServer is used to implement helloworld.GreeterServer.
type busiServer struct {
	UnimplementedBusiServer
}

// SayHello implements helloworld.GreeterServer
func (s *busiServer) Call(ctx context.Context, in *dtmcli.BusiRequest) (*dtmcli.BusiReply, error) {
	log.Printf("busiServer received: %v", in)
	return &dtmcli.BusiReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

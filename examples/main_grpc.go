package examples

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	dtmpb "github.com/yedf/dtm/dtmpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// busiServer is used to implement helloworld.GreeterServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) TransIn(ctx context.Context, in *dtmpb.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.AppData, &req)
	dtmcli.Logf("busiServer %s received: %v %v", dtmcli.GetFuncName(), in.Info, req)
	return &emptypb.Empty{}, nil
}

func (s *busiServer) TransOut(ctx context.Context, in *dtmpb.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.AppData, &req)
	dtmcli.Logf("busiServer %s received: %v %v", dtmcli.GetFuncName(), in.Info, req)
	return &emptypb.Empty{}, nil
}

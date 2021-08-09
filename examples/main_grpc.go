package examples

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	dtmpb "github.com/yedf/dtm/dtmpb"
)

// busiServer is used to implement helloworld.GreeterServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) Call(ctx context.Context, in *dtmpb.BusiRequest) (*dtmpb.BusiReply, error) {
	dtmcli.Logf("busiServer %s received: %v", dtmcli.GetFuncName(), in)
	return &dtmpb.BusiReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

func (s *busiServer) TransIn(ctx context.Context, in *dtmpb.BusiRequest) (*dtmpb.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.AppData, &req)
	dtmcli.Logf("busiServer %s received: %v %v", dtmcli.GetFuncName(), in.Info, req)
	return &dtmpb.BusiReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

func (s *busiServer) TransOut(ctx context.Context, in *dtmpb.BusiRequest) (*dtmpb.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.AppData, &req)
	dtmcli.Logf("busiServer %s received: %v %v", dtmcli.GetFuncName(), in.Info, req)
	return &dtmpb.BusiReply{DtmResult: "SUCCESS", DtmMessage: "ok"}, nil
}

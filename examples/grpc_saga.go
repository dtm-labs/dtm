package examples

import (
	"github.com/yedf/dtm/dtmcli/dtmimp"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

func init() {
	addSample("grpc_saga", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSagaGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		err := saga.Submit()
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
	addSample("grpc_saga_wait", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSagaGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		saga.WaitResult = true
		err := saga.Submit()
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
}

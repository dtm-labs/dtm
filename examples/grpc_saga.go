package examples

import (
	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

func init() {
	addSample("grpc_saga", func() string {
		req := dtmcli.MustMarshal(&TransReq{Amount: 30})
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSaga(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
	addSample("grpc_saga_wait", func() string {
		req := dtmcli.MustMarshal(&TransReq{Amount: 30})
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSaga(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		saga.WaitResult = true
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
}

package examples

import (
	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

func init() {
	addSample("grpc_saga", func() string {
		req := dtmcli.MustMarshal(&TransReq{Amount: 30})
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		msg := dtmgrpc.NewSaga(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOut", BusiGrpc+"/examples.Busi/TransOutRevert", req).
			Add(BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransOutRevert", req)
		err := msg.Submit()
		dtmcli.FatalIfError(err)
		return msg.Gid
	})
}

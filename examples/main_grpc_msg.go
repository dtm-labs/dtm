package examples

import (
	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

// MsgGrpcFireRequest 1
func MsgGrpcFireRequest() string {
	req := dtmcli.MustMarshal(&TransReq{Amount: 30})
	msg := dtmgrpc.NewMsgGrpc(DtmGrpcServer, dtmcli.MustGenGid(DtmServer)).
		Add(BusiGrpc+"/examples.Busi/TransOut", req).
		Add(BusiGrpc+"/examples.Busi/TransIn", req)
	err := msg.Submit()
	e2p(err)
	return msg.Gid
}

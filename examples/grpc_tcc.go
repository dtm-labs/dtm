package examples

import (
	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
)

func init() {
	addSample("grpc_tcc", func() string {
		dtmcli.Logf("tcc simple transaction begin")
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		err := dtmgrpc.TccGlobalTransaction(DtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
			data := dtmcli.MustMarshal(&TransReq{Amount: 30})
			_, err := tcc.CallBranch(data, BusiGrpc+"/examples.Busi/TransOutTcc", BusiGrpc+"/examples.Busi/TransOutConfirm", BusiGrpc+"/examples.Busi/TransOutRevert")
			if err != nil {
				return err
			}
			_, err = tcc.CallBranch(data, BusiGrpc+"/examples.Busi/TransInTcc", BusiGrpc+"/examples.Busi/TransInConfirm", BusiGrpc+"/examples.Busi/TransInRevert")
			return err
		})
		dtmcli.FatalIfError(err)
		return gid
	})
}

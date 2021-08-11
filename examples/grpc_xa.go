package examples

import (
	context "context"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// XaGrpcClient XA client connection
var XaGrpcClient *dtmgrpc.XaGrpcClient = nil

func init() {
	setupFuncs["XaGrpcSetup"] = func(app *gin.Engine) {
		XaGrpcClient = dtmgrpc.NewXaGrpcClient(DtmGrpcServer, config.DB, BusiGrpc+"/examples.Busi/XaNotify")
	}
}

func (s *busiServer) XaNotify(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	err := XaGrpcClient.HandleCallback(in.Info.Gid, in.Info.BranchID, in.Info.BranchType)
	return &emptypb.Empty{}, dtmgrpc.Result2Error(nil, err)
}

// XaGrpcFireRequest 注册全局XA事务，调用XA的分支
func XaGrpcFireRequest() string {
	gid := dtmcli.MustGenGid(DtmServer)
	busiData := dtmcli.MustMarshal(&TransReq{Amount: 30})
	err := XaGrpcClient.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		_, err := xa.CallBranch(busiData, BusiGrpc+"/examples.Busi/TransOutXa")
		if err != nil {
			return err
		}
		_, err = xa.CallBranch(busiData, BusiGrpc+"/examples.Busi/TransInXa")
		return err
	})
	e2p(err)
	return gid
}

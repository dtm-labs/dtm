package examples

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	addSample("grpc_saga_barrier", func() string {
		req := dtmcli.MustMarshal(&TransReq{Amount: 30})
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSaga(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOutBSaga", BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
			Add(BusiGrpc+"/examples.Busi/TransInBSaga", BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req)
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
}

func sagaGrpcBarrierAdjustBalance(sdb *sql.Tx, uid int, amount int) (interface{}, error) {
	_, err := dtmcli.StxExec(sdb, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return dtmcli.ResultSuccess, err

}

func (s *busiServer) TransInBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), req.TransInResult, dtmcli.GetFuncName())
}

func (s *busiServer) TransOutBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), req.TransOutResult, dtmcli.GetFuncName())
}

func (s *busiServer) TransInRevertBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInRevertResult.Fetch(), "", dtmcli.GetFuncName())
}

func (s *busiServer) TransOutRevertBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutRevertResult.Fetch(), "", dtmcli.GetFuncName())
}

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return barrier.Call(sdbGet(), func(sdb *sql.Tx) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 1, req.Amount)
	})
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return barrier.Call(sdbGet(), func(sdb *sql.Tx) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 1, -reqFrom(c).Amount)
	})
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return barrier.Call(sdbGet(), func(sdb *sql.Tx) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 2, -req.Amount)
	})
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return barrier.Call(sdbGet(), func(sdb *sql.Tx) (interface{}, error) {
		return sagaBarrierAdjustBalance(sdb, 2, reqFrom(c).Amount)
	})
}

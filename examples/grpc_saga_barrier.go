package examples

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	addSample("grpc_saga_barrier", func() string {
		req := &BusiReq{Amount: 30}
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSagaGrpc(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOutBSaga", BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
			Add(BusiGrpc+"/examples.Busi/TransInBSaga", BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
		err := saga.Submit()
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
}

func sagaGrpcBarrierAdjustBalance(db dtmcli.DB, uid int, amount int64, result string) error {
	if result == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	}
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err

}

func (s *busiServer) TransInBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(tx dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(tx, 2, in.Amount, in.TransInResult)
	})
}

func (s *busiServer) TransOutBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 1, -in.Amount, in.TransOutResult)
	})
}

func (s *busiServer) TransInRevertBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 2, -in.Amount, "")
	})
}

func (s *busiServer) TransOutRevertBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 1, in.Amount, "")
	})
}

package examples

import (
	"context"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	addSample("grpc_saga_barrier", func() string {
		req := dtmcli.MustMarshal(&TransReq{Amount: 30})
		gid := dtmgrpc.MustGenGid(DtmGrpcServer)
		saga := dtmgrpc.NewSaga(DtmGrpcServer, gid).
			Add(BusiGrpc+"/examples.Busi/TransOutBSaga", BusiGrpc+"/examples.Busi/TransOutRevertBSaga", req).
			Add(BusiGrpc+"/examples.Busi/TransInBSaga", BusiGrpc+"/examples.Busi/TransInRevertBSaga", req)
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
}

func sagaGrpcBarrierAdjustBalance(db dtmcli.DB, uid int, amount int, result string) error {
	if result == dtmcli.ResultFailure {
		return status.New(codes.Aborted, "user rollback").Err()
	}
	_, err := dtmcli.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err

}

func (s *busiServer) TransInBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*dtmgrpc.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	barrier := MustBarrierFromGrpc(in)
	return &dtmgrpc.BusiReply{}, barrier.Call(txGet(), func(tx dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(tx, 2, req.Amount, req.TransInResult)
	})
}

func (s *busiServer) TransOutBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*dtmgrpc.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	barrier := MustBarrierFromGrpc(in)
	return &dtmgrpc.BusiReply{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 1, -req.Amount, req.TransOutResult)
	})
}

func (s *busiServer) TransInRevertBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*dtmgrpc.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	barrier := MustBarrierFromGrpc(in)
	return &dtmgrpc.BusiReply{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 2, -req.Amount, "")
	})
}

func (s *busiServer) TransOutRevertBSaga(ctx context.Context, in *dtmgrpc.BusiRequest) (*dtmgrpc.BusiReply, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	barrier := MustBarrierFromGrpc(in)
	return &dtmgrpc.BusiReply{}, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return sagaGrpcBarrierAdjustBalance(db, 1, req.Amount, "")
	})
}

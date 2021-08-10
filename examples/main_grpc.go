package examples

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/yedf/dtm/dtmcli"
	dtmgrpc "github.com/yedf/dtm/dtmgrpc"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BusiGrpc busi service grpc address
var BusiGrpc string = fmt.Sprintf("localhost:%d", BusiGrpcPort)

// DtmClient grpc client for dtm
var DtmClient dtmgrpc.DtmClient = nil

// GrpcStartup for grpc
func GrpcStartup() {
	conn, err := grpc.Dial(DtmGrpcServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(dtmgrpc.GrpcClientLog))
	dtmcli.FatalIfError(err)
	DtmClient = dtmgrpc.NewDtmClient(conn)
	dtmcli.Logf("dtm client inited")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", BusiGrpcPort))
	dtmcli.FatalIfError(err)
	s := grpc.NewServer(grpc.UnaryInterceptor(dtmgrpc.GrpcServerLog))
	RegisterBusiServer(s, &busiServer{})
	dtmcli.Logf("busi grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		dtmcli.FatalIfError(err)
	}()
}

func handleGrpcBusiness(in *dtmgrpc.BusiRequest, result1 string, result2 string, busi string) error {
	res := dtmcli.OrString(result1, result2, "SUCCESS")
	dtmcli.Logf("grpc busi %s %s result: %s", busi, in.Info, res)
	if res == "SUCCESS" {
		return nil
	} else if res == "FAILURE" {
		return status.New(codes.Aborted, "user want to rollback").Err()
	}
	return status.New(codes.Internal, fmt.Sprintf("unknow result %s", res)).Err()

}

// busiServer is used to implement helloworld.GreeterServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) CanSubmit(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	res := MainSwitch.CanSubmitResult.Fetch()
	return &emptypb.Empty{}, dtmgrpc.Result2Error(res, nil)
}

func (s *busiServer) TransIn(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), req.TransInResult, dtmcli.GetFuncName())
}

func (s *busiServer) TransOut(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), req.TransOutResult, dtmcli.GetFuncName())
}

func (s *busiServer) TransInRevert(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInRevertResult.Fetch(), "", dtmcli.GetFuncName())
}

func (s *busiServer) TransOutRevert(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutRevertResult.Fetch(), "", dtmcli.GetFuncName())
}

func (s *busiServer) TransInConfirm(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInConfirmResult.Fetch(), "", dtmcli.GetFuncName())
}

func (s *busiServer) TransOutConfirm(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutConfirmResult.Fetch(), "", dtmcli.GetFuncName())
}

func (s *busiServer) TransInXa(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if req.TransInResult == "FAILURE" {
			return status.New(codes.Aborted, "user return failure").Err()
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", req.Amount, 2)
		return err
	})
}

func (s *busiServer) TransOutXa(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if req.TransOutResult == "FAILURE" {
			return status.New(codes.Aborted, "user return failure").Err()
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", req.Amount, 1)
		return err
	})
}

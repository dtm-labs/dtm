/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"

	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// BusiGrpc busi service grpc address
var BusiGrpc string = fmt.Sprintf("localhost:%d", BusiGrpcPort)

// DtmClient grpc client for dtm
var DtmClient dtmgimp.DtmClient = nil

// XaGrpcClient XA client connection
var XaGrpcClient *dtmgrpc.XaGrpcClient = nil

func init() {
	setupFuncs["XaGrpcSetup"] = func(app *gin.Engine) {
		XaGrpcClient = dtmgrpc.NewXaGrpcClient(DtmGrpcServer, config.DB, BusiGrpc+"/examples.Busi/XaNotify")
	}
}

// GrpcStartup for grpc
func GrpcStartup() {
	conn, err := grpc.Dial(DtmGrpcServer, grpc.WithInsecure(), grpc.WithUnaryInterceptor(dtmgimp.GrpcClientLog))
	dtmimp.FatalIfError(err)
	DtmClient = dtmgimp.NewDtmClient(conn)
	dtmimp.Logf("dtm client inited")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", BusiGrpcPort))
	dtmimp.FatalIfError(err)
	s := grpc.NewServer(grpc.UnaryInterceptor(dtmgimp.GrpcServerLog))
	RegisterBusiServer(s, &busiServer{})
	go func() {
		dtmimp.Logf("busi grpc listening at %v", lis.Addr())
		err := s.Serve(lis)
		dtmimp.FatalIfError(err)
	}()
	time.Sleep(100 * time.Millisecond)
}

func handleGrpcBusiness(in *BusiReq, result1 string, result2 string, busi string) error {
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	dtmimp.Logf("grpc busi %s %v %s %s result: %s", busi, in, result1, result2, res)
	if res == dtmcli.ResultSuccess {
		return nil
	} else if res == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	} else if res == dtmcli.ResultOngoing {
		return status.New(codes.Aborted, dtmcli.ResultOngoing).Err()
	}
	return status.New(codes.Internal, fmt.Sprintf("unknow result %s", res)).Err()
}

// busiServer is used to implement examples.BusiServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) CanSubmit(ctx context.Context, in *BusiReq) (*BusiReply, error) {
	res := MainSwitch.CanSubmitResult.Fetch()
	return &BusiReply{Message: "a sample"}, dtmgimp.Result2Error(res, nil)
}

func (s *busiServer) TransIn(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOut(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransInRevert(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInRevertResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransOutRevert(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutRevertResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransInConfirm(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInConfirmResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransOutConfirm(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutConfirmResult.Fetch(), "", dtmimp.GetFuncName())
}

func (s *busiServer) TransInTcc(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOutTcc(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransInXa(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(ctx, in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if in.TransInResult == dtmcli.ResultFailure {
			return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
		}
		_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", in.Amount, 2)
		return err
	})
}

func (s *busiServer) TransOutXa(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(ctx, in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if in.TransOutResult == dtmcli.ResultFailure {
			return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
		}
		_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", in.Amount, 1)
		return err
	})
}

func (s *busiServer) TransInTccNested(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	tcc, err := dtmgrpc.TccFromGrpc(ctx)
	dtmimp.FatalIfError(err)
	r := &emptypb.Empty{}
	err = tcc.CallBranch(in, BusiGrpc+"/examples.Busi/TransIn", BusiGrpc+"/examples.Busi/TransInConfirm", BusiGrpc+"/examples.Busi/TransInRevert", r)
	dtmimp.FatalIfError(err)
	return r, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"

	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// BusiGrpc busi service grpc address
var BusiGrpc = fmt.Sprintf("localhost:%d", BusiGrpcPort)

// DtmClient grpc client for dtm
var DtmClient dtmgpb.DtmClient

// XaGrpcClient XA client connection
var XaGrpcClient *dtmgrpc.XaGrpcClient

func init() {
	setupFuncs["XaGrpcSetup"] = func(app *gin.Engine) {
		XaGrpcClient = dtmgrpc.NewXaGrpcClient(dtmutil.DefaultGrpcServer, BusiConf, BusiGrpc+"/busi.Busi/XaNotify")
	}
}

// GrpcStartup for grpc
func GrpcStartup() {
	conn, err := grpc.Dial(dtmutil.DefaultGrpcServer, grpc.WithInsecure(), grpc.WithUnaryInterceptor(dtmgimp.GrpcClientLog))
	logger.FatalIfError(err)
	DtmClient = dtmgpb.NewDtmClient(conn)
	logger.Debugf("dtm client inited")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", BusiGrpcPort))
	logger.FatalIfError(err)
	s := grpc.NewServer(grpc.UnaryInterceptor(dtmgimp.GrpcServerLog))
	RegisterBusiServer(s, &busiServer{})
	go func() {
		logger.Debugf("busi grpc listening at %v", lis.Addr())
		err := s.Serve(lis)
		logger.FatalIfError(err)
	}()
}

// busiServer is used to implement busi.BusiServer.
type busiServer struct {
	UnimplementedBusiServer
}

func (s *busiServer) QueryPrepared(ctx context.Context, in *BusiReq) (*BusiReply, error) {
	res := MainSwitch.QueryPreparedResult.Fetch()
	err := dtmcli.String2DtmError(res)

	return &BusiReply{Message: "a sample data"}, dtmgrpc.DtmError2GrpcError(err)
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
		return sagaGrpcAdjustBalance(db, TransInUID, in.Amount, in.TransInResult)
	})
}

func (s *busiServer) TransOutXa(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(ctx, in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		return sagaGrpcAdjustBalance(db, TransOutUID, in.Amount, in.TransOutResult)
	})
}

func (s *busiServer) TransInTccNested(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	tcc, err := dtmgrpc.TccFromGrpc(ctx)
	logger.FatalIfError(err)
	r := &emptypb.Empty{}
	err = tcc.CallBranch(in, BusiGrpc+"/busi.Busi/TransIn", BusiGrpc+"/busi.Busi/TransInConfirm", BusiGrpc+"/busi.Busi/TransInRevert", r)
	logger.FatalIfError(err)
	return r, handleGrpcBusiness(in, MainSwitch.TransInResult.Fetch(), in.TransInResult, dtmimp.GetFuncName())
}

func (s *busiServer) XaNotify(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return XaGrpcClient.HandleCallback(ctx)
}

func (s *busiServer) TransOutHeaderYes(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	meta := dtmgimp.GetMetaFromContext(ctx, "test_header")
	if meta == "" {
		return &emptypb.Empty{}, errors.New("no header found in HeaderYes")
	}
	return &emptypb.Empty{}, handleGrpcBusiness(in, MainSwitch.TransOutResult.Fetch(), in.TransOutResult, dtmimp.GetFuncName())
}

func (s *busiServer) TransOutHeaderNo(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	meta := dtmgimp.GetMetaFromContext(ctx, "test_header")
	if meta != "" {
		return &emptypb.Empty{}, errors.New("header found in HeaderNo")
	}
	return &emptypb.Empty{}, nil
}

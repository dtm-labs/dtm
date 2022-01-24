/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// XaGrpcGlobalFunc type of xa global function
type XaGrpcGlobalFunc func(xa *XaGrpc) error

// XaGrpcLocalFunc type of xa local function
type XaGrpcLocalFunc func(db *sql.DB, xa *XaGrpc) error

// XaGrpcClient xa client
type XaGrpcClient struct {
	dtmimp.XaClientBase
}

// XaGrpc xa transaction
type XaGrpc struct {
	dtmimp.TransBase
}

// XaGrpcFromRequest construct xa info from request
func XaGrpcFromRequest(ctx context.Context) (*XaGrpc, error) {
	xa := &XaGrpc{
		TransBase: *dtmgimp.TransBaseFromGrpc(ctx),
	}
	if xa.Gid == "" || xa.BranchID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s branchid: %s", xa.Gid, xa.BranchID)
	}
	return xa, nil
}

// NewXaGrpcClient construct a xa client
func NewXaGrpcClient(server string, mysqlConf dtmcli.DBConf, notifyURL string) *XaGrpcClient {
	xa := &XaGrpcClient{XaClientBase: dtmimp.XaClientBase{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}}
	return xa
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaGrpcClient) HandleCallback(ctx context.Context) (*emptypb.Empty, error) {
	tb := dtmgimp.TransBaseFromGrpc(ctx)
	return &emptypb.Empty{}, xc.XaClientBase.HandleCallback(tb.Gid, tb.BranchID, tb.Op)
}

// XaLocalTransaction start a xa local transaction
func (xc *XaGrpcClient) XaLocalTransaction(ctx context.Context, msg proto.Message, xaFunc XaGrpcLocalFunc) error {
	xa, err := XaGrpcFromRequest(ctx)
	if err != nil {
		return err
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return xc.HandleLocalTrans(&xa.TransBase, func(db *sql.DB) error {
		err := xaFunc(db, xa)
		if err != nil {
			return err
		}
		_, err = dtmgimp.MustGetDtmClient(xa.Dtm).RegisterBranch(context.Background(), &dtmgpb.DtmBranchRequest{
			Gid:         xa.Gid,
			BranchID:    xa.BranchID,
			TransType:   xa.TransType,
			BusiPayload: data,
			Data:        map[string]string{"url": xc.NotifyURL},
		})
		return err
	})
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaGrpcClient) XaGlobalTransaction(gid string, xaFunc XaGrpcGlobalFunc) error {
	return xc.XaGlobalTransaction2(gid, func(xg *XaGrpc) {}, xaFunc)
}

// XaGlobalTransaction2 new version of XaGlobalTransaction. support custom
func (xc *XaGrpcClient) XaGlobalTransaction2(gid string, custom func(*XaGrpc), xaFunc XaGrpcGlobalFunc) error {
	xa := &XaGrpc{TransBase: *dtmimp.NewTransBase(gid, "xa", xc.Server, "")}
	custom(xa)
	dc := dtmgimp.MustGetDtmClient(xa.Dtm)
	req := &dtmgpb.DtmRequest{
		Gid:       gid,
		TransType: xa.TransType,
	}
	return xc.HandleGlobalTrans(&xa.TransBase, func(action string) error {
		f := map[string]func(context.Context, *dtmgpb.DtmRequest, ...grpc.CallOption) (*emptypb.Empty, error){
			"prepare": dc.Prepare,
			"submit":  dc.Submit,
			"abort":   dc.Abort,
		}[action]
		_, err := f(context.Background(), req)
		return err
	}, func() error {
		return xaFunc(xa)
	})
}

// CallBranch call a xa branch
func (x *XaGrpc) CallBranch(msg proto.Message, url string, reply interface{}) error {
	return dtmgimp.InvokeBranch(&x.TransBase, false, msg, url, reply, x.NewSubBranchID(), "action")
}

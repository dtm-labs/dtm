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

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// XaGrpcGlobalFunc type of xa global function
type XaGrpcGlobalFunc func(xa *XaGrpc) error

// XaGrpcLocalFunc type of xa local function
type XaGrpcLocalFunc func(db *sql.DB, xa *XaGrpc) error

// XaGrpc xa transaction
type XaGrpc struct {
	dtmimp.TransBase
	Phase2URL string
}

// XaGrpcFromRequest construct xa info from request
func XaGrpcFromRequest(ctx context.Context) (*XaGrpc, error) {
	xa := &XaGrpc{
		TransBase: *dtmgimp.TransBaseFromGrpc(ctx),
	}
	xa.Phase2URL = dtmgimp.GetDtmMetaFromContext(ctx, "phase2_url")
	if xa.Gid == "" || xa.BranchID == "" || xa.Op == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s branchid: %s op: %s phase2_url: %s", xa.Gid, xa.BranchID, xa.Op, xa.Phase2URL)
	}
	return xa, nil
}

// XaLocalTransaction start a xa local transaction
func XaLocalTransaction(ctx context.Context, dbConf dtmcli.DBConf, xaFunc XaGrpcLocalFunc) error {
	xa, err := XaGrpcFromRequest(ctx)
	if err != nil {
		return err
	}
	if xa.Op == dtmimp.OpCommit || xa.Op == dtmimp.OpRollback {
		return dtmimp.XaHandlePhase2(xa.Gid, dbConf, xa.BranchID, xa.Op)
	}
	return dtmimp.XaHandleLocalTrans(&xa.TransBase, dbConf, func(db *sql.DB) error {
		err := xaFunc(db, xa)
		if err != nil {
			return err
		}
		_, err = dtmgimp.MustGetDtmClient(xa.Dtm).RegisterBranch(context.Background(), &dtmgpb.DtmBranchRequest{
			Gid:         xa.Gid,
			BranchID:    xa.BranchID,
			TransType:   xa.TransType,
			BusiPayload: nil,
			Data:        map[string]string{"url": xa.Phase2URL},
		})
		return err
	})
}

// XaGlobalTransaction start a xa global transaction
func XaGlobalTransaction(server string, gid string, xaFunc XaGrpcGlobalFunc) error {
	return XaGlobalTransaction2(server, gid, func(xg *XaGrpc) {}, xaFunc)
}

// XaGlobalTransaction2 new version of XaGlobalTransaction. support custom
func XaGlobalTransaction2(server string, gid string, custom func(*XaGrpc), xaFunc XaGrpcGlobalFunc) error {
	xa := &XaGrpc{TransBase: *dtmimp.NewTransBase(gid, "xa", server, "")}
	custom(xa)
	dc := dtmgimp.MustGetDtmClient(xa.Dtm)
	req := dtmgimp.GetDtmRequest(&xa.TransBase)
	return dtmimp.XaHandleGlobalTrans(&xa.TransBase, func(action string) error {
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
func (x *XaGrpc) CallBranch(msg proto.Message, url string, reply interface{}, opts ...grpc.CallOption) error {
	return dtmgimp.InvokeBranch(&x.TransBase, false, msg, url, reply, x.NewSubBranchID(), "action", opts...)
}

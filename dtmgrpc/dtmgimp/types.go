/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgimp

import (
	"context"
	"fmt"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtmdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// GrpcServerLog middleware to print server-side grpc log
func GrpcServerLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	began := time.Now()
	logger.Debugf("grpc server handling: %s %s", info.FullMethod, dtmimp.MustMarshalString(req))
	LogDtmCtx(ctx)
	m, err := handler(ctx, req)
	res := fmt.Sprintf("%2dms %v %s %s %s",
		time.Since(began).Milliseconds(), err, info.FullMethod, dtmimp.MustMarshalString(m), dtmimp.MustMarshalString(req))
	st, _ := status.FromError(err)
	if err == nil || st != nil && st.Code() == codes.FailedPrecondition {
		logger.Infof("%s", res)
	} else {
		logger.Errorf("%s", res)
	}
	return m, err
}

// GrpcClientLog middleware to print client-side grpc log
func GrpcClientLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger.Debugf("grpc client calling: %s%s %v", cc.Target(), method, dtmimp.MustMarshalString(req))
	LogDtmCtx(ctx)
	err := invoker(ctx, method, req, reply, cc, opts...)
	res := fmt.Sprintf("grpc client called: %s%s %s result: %s err: %v",
		cc.Target(), method, dtmimp.MustMarshalString(req), dtmimp.MustMarshalString(reply), err)
	st, _ := status.FromError(err)
	if err == nil || st != nil && st.Code() == codes.FailedPrecondition {
		logger.Infof("%s", res)
	} else {
		logger.Errorf("%s", res)
	}
	return err
}

// InvokeBranch invoke a url for trans
func InvokeBranch(t *dtmimp.TransBase, isRaw bool, msg proto.Message, url string, reply interface{}, branchID string, op string) error {
	server, method, err := dtmdriver.GetDriver().ParseServerMethod(url)
	if err != nil {
		return err
	}
	ctx := TransInfo2Ctx(t.Context, t.Gid, t.TransType, branchID, op, t.Dtm)
	ctx = metadata.AppendToOutgoingContext(ctx, Map2Kvs(t.BranchHeaders)...)
	if t.TransType == "xa" { // xa branch need additional phase2_url
		ctx = metadata.AppendToOutgoingContext(ctx, Map2Kvs(map[string]string{dtmpre + "phase2_url": url})...)
	}
	return MustGetGrpcConn(server, isRaw).Invoke(ctx, method, msg, reply)
}

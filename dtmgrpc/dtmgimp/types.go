/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgimp

import (
	"context"
	"fmt"
	"strings"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetServerAndMethod 将grpc的url分解为server和method
func GetServerAndMethod(grpcURL string) (string, string) {
	fs := strings.Split(grpcURL, "/")
	server := fs[0]
	method := "/" + strings.Join(fs[1:], "/")
	return server, method
}

// GrpcServerLog 打印grpc服务端的日志
func GrpcServerLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	dtmimp.Logf("grpc server handling: %s %v", info.FullMethod, req)
	LogDtmCtx(ctx)
	m, err := handler(ctx, req)
	res := fmt.Sprintf("grpc server handled: %s %v result: %v err: %v", info.FullMethod, req, m, err)
	if err != nil {
		dtmimp.LogRedf("%s", res)
	} else {
		dtmimp.Logf("%s", res)
	}
	return m, err
}

// GrpcClientLog 打印grpc服务端的日志
func GrpcClientLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	dtmimp.Logf("grpc client calling: %s%s %v", cc.Target(), method, req)
	LogDtmCtx(ctx)
	err := invoker(ctx, method, req, reply, cc, opts...)
	res := fmt.Sprintf("grpc client called: %s%s %v result: %v err: %v", cc.Target(), method, req, reply, err)
	if err != nil {
		dtmimp.LogRedf("%s", res)
	} else {
		dtmimp.Logf("%s", res)
	}
	return err
}

// Result2Error 将通用的result转成grpc的error
func Result2Error(res interface{}, err error) error {
	e := dtmimp.CheckResult(res, err)
	if e == dtmimp.ErrFailure {
		dtmimp.LogRedf("failure: res: %v, err: %v", res, e)
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	} else if e == dtmimp.ErrOngoing {
		return status.New(codes.Aborted, dtmcli.ResultOngoing).Err()
	}
	return e
}

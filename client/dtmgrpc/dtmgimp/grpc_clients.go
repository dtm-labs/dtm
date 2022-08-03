/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgimp

import (
	"fmt"
	"sync"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type rawCodec struct{}

func (cb rawCodec) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (cb rawCodec) Unmarshal(data []byte, v interface{}) error {
	ba, ok := v.(*[]byte)
	dtmimp.PanicIf(!ok, fmt.Errorf("please pass in *[]byte"))
	*ba = append(*ba, data...)

	return nil
}

func (cb rawCodec) Name() string { return "dtm_raw" }

var normalClients, rawClients sync.Map

// ClientInterceptors declares grpc.UnaryClientInterceptors slice
var ClientInterceptors = []grpc.UnaryClientInterceptor{}

// MustGetDtmClient 1
func MustGetDtmClient(grpcServer string) dtmgpb.DtmClient {
	return dtmgpb.NewDtmClient(MustGetGrpcConn(grpcServer, false))
}

// GetGrpcConn 1
func GetGrpcConn(grpcServer string, isRaw bool) (conn *grpc.ClientConn, rerr error) {
	clients := &normalClients
	if isRaw {
		clients = &rawClients
	}
	grpcServer = dtmimp.MayReplaceLocalhost(grpcServer)
	v, ok := clients.Load(grpcServer)
	if !ok {
		opts := grpc.WithDefaultCallOptions()
		if isRaw {
			opts = grpc.WithDefaultCallOptions(grpc.ForceCodec(rawCodec{}))
		}
		logger.Debugf("grpc client connecting %s", grpcServer)
		interceptors := append(ClientInterceptors, GrpcClientLog)
		interceptors = append(interceptors, dtmdriver.Middlewares.Grpc...)
		inOpt := grpc.WithChainUnaryInterceptor(interceptors...)
		conn, rerr := grpc.Dial(grpcServer, inOpt, grpc.WithTransportCredentials(insecure.NewCredentials()), opts)
		if rerr == nil {
			clients.Store(grpcServer, conn)
			v = conn
			logger.Debugf("grpc client inited for %s", grpcServer)
		}
	}
	return v.(*grpc.ClientConn), rerr
}

// MustGetGrpcConn 1
func MustGetGrpcConn(grpcServer string, isRaw bool) *grpc.ClientConn {
	conn, err := GetGrpcConn(grpcServer, isRaw)
	dtmimp.E2P(err)
	return conn
}

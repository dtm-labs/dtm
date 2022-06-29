/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgimp

import (
	context "context"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// MustProtoMarshal must version of proto.Marshal
func MustProtoMarshal(msg proto.Message) []byte {
	b, err := proto.Marshal(msg)
	dtmimp.PanicIf(err != nil, err)
	return b
}

// MustProtoUnmarshal must version of proto.Unmarshal
func MustProtoUnmarshal(data []byte, msg proto.Message) {
	err := proto.Unmarshal(data, msg)
	dtmimp.PanicIf(err != nil, err)
}

// GetDtmRequest return a DtmRequest from TransBase
func GetDtmRequest(s *dtmimp.TransBase) *dtmgpb.DtmRequest {
	return &dtmgpb.DtmRequest{
		Gid:       s.Gid,
		TransType: s.TransType,
		TransOptions: &dtmgpb.DtmTransOptions{
			WaitResult:         s.WaitResult,
			TimeoutToFail:      s.TimeoutToFail,
			RetryInterval:      s.RetryInterval,
			PassthroughHeaders: s.PassthroughHeaders,
			BranchHeaders:      s.BranchHeaders,
			RequestTimeout:     s.RequestTimeout,
			RollbackReason:     s.RollbackReason,
		},
		QueryPrepared: s.QueryPrepared,
		CustomedData:  s.CustomData,
		BinPayloads:   s.BinPayloads,
		Steps:         dtmimp.MustMarshalString(s.Steps),
	}
}

// DtmGrpcCall make a convenient call to dtm
func DtmGrpcCall(s *dtmimp.TransBase, operation string) error {
	reply := emptypb.Empty{}
	return MustGetGrpcConn(s.Dtm, false).Invoke(s.Context, "/dtmgimp.Dtm/"+operation, GetDtmRequest(s), &reply)
}

const dtmpre string = "dtm-"

// TransInfo2Ctx add trans info to grpc context
func TransInfo2Ctx(ctx context.Context, gid, transType, branchID, op, dtm string) context.Context {
	md := metadata.Pairs(
		dtmpre+"gid", gid,
		dtmpre+"trans_type", transType,
		dtmpre+"branch_id", branchID,
		dtmpre+"op", op,
		dtmpre+"dtm", dtm,
	)
	nctx := ctx
	if ctx == nil {
		nctx = context.Background()
	}
	return metadata.NewOutgoingContext(nctx, md)
}

// Map2Kvs map to metadata kv
func Map2Kvs(m map[string]string) []string {
	kvs := []string{}
	for k, v := range m {
		kvs = append(kvs, k, v)
	}
	return kvs
}

// LogDtmCtx logout dtm info in context metadata
func LogDtmCtx(ctx context.Context) {
	tb := TransBaseFromGrpc(ctx)
	if tb.Gid != "" {
		logger.Debugf("gid: %s trans_type: %s branch_id: %s op: %s dtm: %s", tb.Gid, tb.TransType, tb.BranchID, tb.Op, tb.Dtm)
	}
}

func dtmGet(md metadata.MD, key string) string {
	return mdGet(md, dtmpre+key)
}

func mdGet(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// TransBaseFromGrpc get trans base info from a context metadata
func TransBaseFromGrpc(ctx context.Context) *dtmimp.TransBase {
	md, _ := metadata.FromIncomingContext(ctx)
	tb := dtmimp.NewTransBase(dtmGet(md, "gid"), dtmGet(md, "trans_type"), dtmGet(md, "dtm"), dtmGet(md, "branch_id"))
	tb.Op = dtmGet(md, "op")
	return tb
}

// GetMetaFromContext get header from context
func GetMetaFromContext(ctx context.Context, name string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	return mdGet(md, name)
}

// GetDtmMetaFromContext get dtm header from context
func GetDtmMetaFromContext(ctx context.Context, name string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	return dtmGet(md, name)
}

type requestTimeoutKey struct{}

// RequestTimeoutFromContext returns requestTime of transOption option
func RequestTimeoutFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(requestTimeoutKey{}).(int64); ok {
		return v
	}

	return 0
}

// RequestTimeoutNewContext sets requestTimeout of transOption option to context
func RequestTimeoutNewContext(ctx context.Context, requestTimeout int64) context.Context {
	return context.WithValue(ctx, requestTimeoutKey{}, requestTimeout)
}

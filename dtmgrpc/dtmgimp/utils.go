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

// DtmGrpcCall make a convenient call to dtm
func DtmGrpcCall(s *dtmimp.TransBase, operation string) error {
	reply := emptypb.Empty{}
	return MustGetGrpcConn(s.Dtm, false).Invoke(context.Background(), "/dtmgimp.Dtm/"+operation, &dtmgpb.DtmRequest{
		Gid:       s.Gid,
		TransType: s.TransType,
		TransOptions: &dtmgpb.DtmTransOptions{
			WaitResult:         s.WaitResult,
			TimeoutToFail:      s.TimeoutToFail,
			RetryInterval:      s.RetryInterval,
			PassthroughHeaders: s.PassthroughHeaders,
			BranchHeaders:      s.BranchHeaders,
		},
		QueryPrepared: s.QueryPrepared,
		CustomedData:  s.CustomData,
		BinPayloads:   s.BinPayloads,
		Steps:         dtmimp.MustMarshalString(s.Steps),
	}, &reply)
}

const dtmpre string = "dtm-"

// TransInfo2Ctx add trans info to grpc context
func TransInfo2Ctx(gid, transType, branchID, op, dtm string) context.Context {
	md := metadata.Pairs(
		dtmpre+"gid", gid,
		dtmpre+"trans_type", transType,
		dtmpre+"branch_id", branchID,
		dtmpre+"op", op,
		dtmpre+"dtm", dtm,
	)
	return metadata.NewOutgoingContext(context.Background(), md)
}

// Map2Kvs map to metadata kv
func Map2Kvs(m map[string]string) []string {
	var kvs []string
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

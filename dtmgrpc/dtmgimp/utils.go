/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgimp

import (
	context "context"

	"github.com/yedf/dtm/dtmcli/dtmimp"
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
	return MustGetGrpcConn(s.Dtm, false).Invoke(context.Background(), "/dtmgimp.Dtm/"+operation, &DtmRequest{
		Gid:       s.Gid,
		TransType: s.TransType,
		TransOptions: &DtmTransOptions{
			WaitResult:    s.WaitResult,
			TimeoutToFail: s.TimeoutToFail,
			RetryInterval: s.RetryInterval,
		},
		QueryPrepared: s.QueryPrepared,
		CustomedData:  s.CustomData,
		BinPayloads:   s.BinPayloads,
		Steps:         dtmimp.MustMarshalString(s.Steps),
	}, &reply)
}

const mdpre string = "dtm-"

// TransInfo2Ctx add trans info to grpc context
func TransInfo2Ctx(gid, transType, branchID, op, dtm string) context.Context {
	md := metadata.Pairs(
		mdpre+"gid", gid,
		mdpre+"trans_type", transType,
		mdpre+"branch_id", branchID,
		mdpre+"op", op,
		mdpre+"dtm", dtm,
	)
	return metadata.NewOutgoingContext(context.Background(), md)
}

// LogDtmCtx logout dtm info in context metadata
func LogDtmCtx(ctx context.Context) {
	tb := TransBaseFromGrpc(ctx)
	if tb.Gid != "" {
		dtmimp.Logf("gid: %s trans_type: %s branch_id: %s op: %s dtm: %s", tb.Gid, tb.TransType, tb.BranchID, tb.Op, tb.Dtm)
	}
}

func mdGet(md metadata.MD, key string) string {
	v := md.Get(mdpre + key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// TransBaseFromGrpc get trans base info from a context metadata
func TransBaseFromGrpc(ctx context.Context) *dtmimp.TransBase {
	md, _ := metadata.FromIncomingContext(ctx)
	tb := dtmimp.NewTransBase(mdGet(md, "gid"), mdGet(md, "trans_type"), mdGet(md, "dtm"), mdGet(md, "branch_id"))
	tb.Op = mdGet(md, "op")
	return tb
}

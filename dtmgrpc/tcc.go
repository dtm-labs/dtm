/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	context "context"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"google.golang.org/protobuf/proto"
)

// TccGrpc struct of tcc
type TccGrpc struct {
	dtmimp.TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *TccGrpc) error

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm服务器地址
// gid 全局事务id
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	return TccGlobalTransaction2(dtm, gid, func(tg *TccGrpc) {}, tccFunc)
}

// TccGlobalTransaction2 new version of TccGlobalTransaction
func TccGlobalTransaction2(dtm string, gid string, custom func(*TccGrpc), tccFunc TccGlobalFunc) (rerr error) {
	tcc := &TccGrpc{TransBase: *dtmimp.NewTransBase(gid, "tcc", dtm, "")}
	custom(tcc)
	rerr = dtmgimp.DtmGrpcCall(&tcc.TransBase, "Prepare")
	if rerr != nil {
		return rerr
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer dtmimp.DeferDo(&rerr, func() error {
		return dtmgimp.DtmGrpcCall(&tcc.TransBase, "Submit")
	}, func() error {
		return dtmgimp.DtmGrpcCall(&tcc.TransBase, "Abort")
	})
	return tccFunc(tcc)
}

// TccFromGrpc tcc from request info
func TccFromGrpc(ctx context.Context) (*TccGrpc, error) {
	tcc := &TccGrpc{
		TransBase: *dtmgimp.TransBaseFromGrpc(ctx),
	}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s branchid: %s", tcc.Dtm, tcc.Gid, tcc.BranchID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
func (t *TccGrpc) CallBranch(busiMsg proto.Message, tryURL string, confirmURL string, cancelURL string, reply interface{}) error {
	branchID := t.NewSubBranchID()
	bd, err := proto.Marshal(busiMsg)
	if err == nil {
		_, err = dtmgimp.MustGetDtmClient(t.Dtm).RegisterBranch(context.Background(), &dtmgpb.DtmBranchRequest{
			Gid:         t.Gid,
			TransType:   t.TransType,
			BranchID:    branchID,
			BusiPayload: bd,
			Data:        map[string]string{"confirm": confirmURL, "cancel": cancelURL},
		})
	}
	if err != nil {
		return err
	}
	return dtmgimp.InvokeBranch(&t.TransBase, false, busiMsg, tryURL, reply, branchID, "try")
}

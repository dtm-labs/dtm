/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func getXcg() *dtmgrpc.XaGrpcClient {
	return busi.XaGrpcClient
}
func TestXaGrpcNormal(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXcg().XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := busi.GenBusiReq(30, false, false)
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		return xa.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransInXa", r)
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestXaGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXcg().XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := busi.GenBusiReq(30, false, true)
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		return xa.CallBranch(req, busi.BusiGrpc+"/busi.Busi/TransInXa", r)
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestXaGrpcType(t *testing.T) {
	gid := dtmimp.GetFuncName()
	_, err := dtmgrpc.XaGrpcFromRequest(context.Background())
	assert.Error(t, err)

	err = busi.XaGrpcClient.XaLocalTransaction(context.Background(), nil, nil)
	assert.Error(t, err)

	err = dtmimp.CatchP(func() {
		busi.XaGrpcClient.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error { panic(fmt.Errorf("hello")) })
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
}

func TestXaGrpcLocalError(t *testing.T) {
	gid := dtmimp.GetFuncName()
	xc := busi.XaGrpcClient
	err := xc.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		return fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
	waitTransProcessed(gid)
}

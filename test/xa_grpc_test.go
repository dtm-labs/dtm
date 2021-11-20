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

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc"
	"github.com/yedf/dtm/examples"
	"google.golang.org/protobuf/types/known/emptypb"
)

func getXcg() *dtmgrpc.XaGrpcClient {
	return examples.XaGrpcClient
}
func TestXaGrpcNormal(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXcg().XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := examples.GenBusiReq(30, false, false)
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		return xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa", r)
	})
	assert.Equal(t, nil, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestXaGrpcRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := getXcg().XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
		req := examples.GenBusiReq(30, false, true)
		r := &emptypb.Empty{}
		err := xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransOutXa", r)
		if err != nil {
			return err
		}
		return xa.CallBranch(req, examples.BusiGrpc+"/examples.Busi/TransInXa", r)
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestXaGrpcType(t *testing.T) {
	_, err := dtmgrpc.XaGrpcFromRequest(context.Background())
	assert.Error(t, err)

	err = examples.XaGrpcClient.XaLocalTransaction(context.Background(), nil, nil)
	assert.Error(t, err)

	err = dtmimp.CatchP(func() {
		examples.XaGrpcClient.XaGlobalTransaction("id1", func(xa *dtmgrpc.XaGrpc) error { panic(fmt.Errorf("hello")) })
	})
	assert.Error(t, err)
}

func TestXaGrpcLocalError(t *testing.T) {
	xc := examples.XaGrpcClient
	err := xc.XaGlobalTransaction(dtmimp.GetFuncName(), func(xa *dtmgrpc.XaGrpc) error {
		return fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
}

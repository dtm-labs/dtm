/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestSagaBarrierNormal(t *testing.T) {
	saga := genSagaBarrier(dtmimp.GetFuncName(), false, false)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaBarrierRollback(t *testing.T) {
	saga := genSagaBarrier(dtmimp.GetFuncName(), false, true)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func genSagaBarrier(gid string, outFailed, inFailed bool) *dtmcli.Saga {
	req := busi.GenTransReq(30, outFailed, inFailed)
	return dtmcli.NewSaga(DtmServer, gid).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCom", req).
		Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCom", req)
}

func TestSagaBarrier2Normal(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	saga := dtmcli.NewSaga(DtmServer, gid).
		Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCom", req).
		Add(Busi+"/SagaB2TransIn", Busi+"/SagaB2TransInCom", req)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

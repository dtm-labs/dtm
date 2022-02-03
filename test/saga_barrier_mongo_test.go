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

func TestSagaBarrierMongoNormal(t *testing.T) {
	before := getBeforeBalances("mongo")
	saga := genSagaBarrierMongo(dtmimp.GetFuncName(), false)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assertNotSameBalance(t, before, "mongo")
}

func TestSagaBarrierMongoRollback(t *testing.T) {
	before := getBeforeBalances("mongo")
	saga := genSagaBarrierMongo(dtmimp.GetFuncName(), true)
	err := saga.Submit()
	assert.Nil(t, err)
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
	assertSameBalance(t, before, "mongo")
}

func genSagaBarrierMongo(gid string, transInFailed bool) *dtmcli.Saga {
	req := busi.GenTransReq(30, false, transInFailed)
	req.Store = "mongo"
	return dtmcli.NewSaga(DtmServer, gid).
		Add(Busi+"/SagaMongoTransOut", Busi+"/SagaMongoTransOutCom", req).
		Add(Busi+"/SagaMongoTransIn", Busi+"/SagaMongoTransInCom", req)
}

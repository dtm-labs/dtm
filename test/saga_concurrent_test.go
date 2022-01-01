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

func genSagaCon(gid string, outFailed bool, inFailed bool) *dtmcli.Saga {
	return genSaga(gid, outFailed, inFailed).EnableConcurrent()
}

func TestSagaConNormal(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), false, false)
	sagaCon.Submit()
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(sagaCon.Gid))
}

func TestSagaConRollbackNormal(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), true, false)
	busi.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	err := sagaCon.Submit()
	assert.Nil(t, err)
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, StatusAborting, getTransStatus(sagaCon.Gid))
	g := cronTransOnce()
	assert.Equal(t, sagaCon.Gid, g)
	assert.Equal(t, StatusFailed, getTransStatus(sagaCon.Gid))
	// TODO should fix this
	// assert.Equal(t, []string{StatusSucceed, StatusFailed, StatusSucceed, StatusSucceed}, getBranchesStatus(sagaCon.Gid))
}

func TestSagaConRollbackOrder(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), true, false)
	sagaCon.AddBranchOrder(1, []int{0})
	err := sagaCon.Submit()
	assert.Nil(t, err)
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, StatusFailed, getTransStatus(sagaCon.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusFailed, StatusPrepared, StatusPrepared}, getBranchesStatus(sagaCon.Gid))
}

func TestSagaConCommittedOngoing(t *testing.T) {
	sagaCon := genSagaCon(dtmimp.GetFuncName(), false, false)
	busi.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	sagaCon.Submit()
	waitTransProcessed(sagaCon.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(sagaCon.Gid))

	g := cronTransOnce()
	assert.Equal(t, sagaCon.Gid, g)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(sagaCon.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(sagaCon.Gid))
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

var conf = &config.Config

func dbGet() *dtmutil.DB {
	return dtmutil.DbGet(busi.BusiConf)
}

// waitTransProcessed only for test usage. wait for transaction processed once
func waitTransProcessed(gid string) {
	logger.Debugf("waiting for gid %s", gid)
	select {
	case id := <-dtmsvr.TransProcessedTestChan:
		logger.FatalfIf(id != gid, "------- expecting: %s but %s found", gid, id)
		logger.Debugf("finish for gid %s", gid)
	case <-time.After(time.Duration(time.Second * 4)):
		logger.FatalfIf(true, "Wait Trans timeout")
	}
}

func cronTransOnce(t *testing.T, gid string) {
	gid2 := dtmsvr.CronTransOnce()
	assert.Equal(t, gid, gid2)
	if dtmsvr.TransProcessedTestChan != nil && gid != "" {
		waitTransProcessed(gid)
	}
}

var e2p = dtmimp.E2P

// TransGlobal alias
type TransGlobal = dtmsvr.TransGlobal

// TransBranch alias
type TransBranch = dtmsvr.TransBranch

func cronTransOnceForwardNow(t *testing.T, gid string, seconds int) {
	old := dtmsvr.NowForwardDuration
	dtmsvr.NowForwardDuration = time.Duration(seconds) * time.Second
	cronTransOnce(t, gid)
	dtmsvr.NowForwardDuration = old
}

func cronTransOnceForwardCron(t *testing.T, gid string, seconds int) {
	old := dtmsvr.CronForwardDuration
	dtmsvr.CronForwardDuration = time.Duration(seconds) * time.Second
	cronTransOnce(t, gid)
	dtmsvr.CronForwardDuration = old
}

const (
	// StatusPrepared status for global/branch trans status.
	StatusPrepared = dtmcli.StatusPrepared
	// StatusSubmitted status for global trans status.
	StatusSubmitted = dtmcli.StatusSubmitted
	// StatusSucceed status for global/branch trans status.
	StatusSucceed = dtmcli.StatusSucceed
	// StatusFailed status for global/branch trans status.
	StatusFailed = dtmcli.StatusFailed
	// StatusAborting status for global trans status.
	StatusAborting = dtmcli.StatusAborting
)

func getBeforeBalances(store string) []int {
	b1 := busi.GetBalanceByUID(busi.TransOutUID, store)
	b2 := busi.GetBalanceByUID(busi.TransInUID, store)
	return []int{b1, b2}
}

func assertSameBalance(t *testing.T, before []int, store string) {
	b1 := busi.GetBalanceByUID(busi.TransOutUID, store)
	b2 := busi.GetBalanceByUID(busi.TransInUID, store)
	assert.Equal(t, before[0], b1)
	assert.Equal(t, before[1], b2)
}

func assertNotSameBalance(t *testing.T, before []int, store string) {
	b1 := busi.GetBalanceByUID(busi.TransOutUID, store)
	b2 := busi.GetBalanceByUID(busi.TransInUID, store)
	assert.NotEqual(t, before[0], b1)
	assert.Equal(t, before[0]+before[1], b1+b2)
}

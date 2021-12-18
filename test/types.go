/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmsvr"
)

var config = &common.Config

func dbGet() *common.DB {
	return common.DbGet(config.ExamplesDB)
}

// waitTransProcessed only for test usage. wait for transaction processed once
func waitTransProcessed(gid string) {
	dtmimp.Logf("waiting for gid %s", gid)
	select {
	case id := <-dtmsvr.TransProcessedTestChan:
		for id != gid {
			dtmimp.LogRedf("-------id %s not match gid %s", id, gid)
			id = <-dtmsvr.TransProcessedTestChan
		}
		dtmimp.Logf("finish for gid %s", gid)
	case <-time.After(time.Duration(time.Second * 3)):
		dtmimp.LogFatalf("Wait Trans timeout")
	}
}

func cronTransOnce() {
	gid := dtmsvr.CronTransOnce()
	if dtmsvr.TransProcessedTestChan != nil && gid != "" {
		waitTransProcessed(gid)
	}
}

var e2p = dtmimp.E2P

// TransGlobal alias
type TransGlobal = dtmsvr.TransGlobal

// TransBranch alias
type TransBranch = dtmsvr.TransBranch

func cronTransOnceForwardNow(seconds int) {
	old := dtmsvr.NowForwardDuration
	dtmsvr.NowForwardDuration = time.Duration(seconds) * time.Second
	cronTransOnce()
	dtmsvr.NowForwardDuration = old
}

func cronTransOnceForwardCron(seconds int) {
	old := dtmsvr.CronForwardDuration
	dtmsvr.CronForwardDuration = time.Duration(seconds) * time.Second
	cronTransOnce()
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

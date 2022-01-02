/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
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
	case <-time.After(time.Duration(time.Second * 3)):
		logger.FatalfIf(true, "Wait Trans timeout")
	}
}

func cronTransOnce() string {
	gid := dtmsvr.CronTransOnce()
	if dtmsvr.TransProcessedTestChan != nil && gid != "" {
		waitTransProcessed(gid)
	}
	return gid
}

var e2p = dtmimp.E2P

// TransGlobal alias
type TransGlobal = dtmsvr.TransGlobal

// TransBranch alias
type TransBranch = dtmsvr.TransBranch

func cronTransOnceForwardNow(seconds int) string {
	old := dtmsvr.NowForwardDuration
	dtmsvr.NowForwardDuration = time.Duration(seconds) * time.Second
	gid := cronTransOnce()
	dtmsvr.NowForwardDuration = old
	return gid
}

func cronTransOnceForwardCron(seconds int) string {
	old := dtmsvr.CronForwardDuration
	dtmsvr.CronForwardDuration = time.Duration(seconds) * time.Second
	gid := cronTransOnce()
	dtmsvr.CronForwardDuration = old
	return gid
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

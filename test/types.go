package test

import (
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
)

var config = common.DtmConfig

func dbGet() *common.DB {
	return common.DbGet(config.DB)
}

// WaitTransProcessed only for test usage. wait for transaction processed once
func WaitTransProcessed(gid string) {
	dtmcli.Logf("waiting for gid %s", gid)
	select {
	case id := <-dtmsvr.TransProcessedTestChan:
		for id != gid {
			dtmcli.LogRedf("-------id %s not match gid %s", id, gid)
			id = <-dtmsvr.TransProcessedTestChan
		}
		dtmcli.Logf("finish for gid %s", gid)
	case <-time.After(time.Duration(time.Second * 3)):
		dtmcli.LogFatalf("Wait Trans timeout")
	}
}

func cronTransOnce() {
	gid := dtmsvr.CronTransOnce()
	if dtmsvr.TransProcessedTestChan != nil && gid != "" {
		WaitTransProcessed(gid)
	}
}

var e2p = dtmcli.E2P

// TransGlobal alias
type TransGlobal = dtmsvr.TransGlobal

// TransBranch alias
type TransBranch = dtmsvr.TransBranch

// M alias
type M = dtmcli.M

func cronTransOnceForwardNow(seconds int) {
	old := dtmsvr.NowForwardDuration
	dtmsvr.NowForwardDuration = time.Duration(seconds) * time.Second
	cronTransOnce()
	dtmsvr.NowForwardDuration = old
}

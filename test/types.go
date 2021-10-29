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

// WaitTransProcessed alias
var WaitTransProcessed = dtmsvr.WaitTransProcessed

// CronTransOnce alias
var CronTransOnce = dtmsvr.CronTransOnce
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
	CronTransOnce()
	dtmsvr.NowForwardDuration = old
}

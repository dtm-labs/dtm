/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
)

// NowForwardDuration will be set in test, trans may be timeout
var NowForwardDuration = time.Duration(0)

// CronForwardDuration will be set in test. cron will fetch trans which expire in CronForwardDuration
var CronForwardDuration = time.Duration(0)

// CronTransOnce cron expired trans. use expireIn as expire time
func CronTransOnce() (gid string) {
	defer handlePanic(nil)
	trans := lockOneTrans(CronForwardDuration)
	if trans == nil {
		return
	}
	gid = trans.Gid
	trans.WaitResult = true
	branches := GetStore().FindBranches(gid)
	err := trans.Process(branches)
	dtmimp.PanicIf(err != nil && !errors.Is(err, dtmcli.ErrFailure) && !errors.Is(err, dtmcli.ErrOngoing), err)
	return
}

// CronExpiredTrans cron expired trans, num == -1 indicate for ever
func CronExpiredTrans(num int) {
	for i := 0; i < num || num == -1; i++ {
		gid := CronTransOnce()
		if gid == "" && num != 1 {
			sleepCronTime()
		}
	}
}

func lockOneTrans(expireIn time.Duration) *TransGlobal {
	global := GetStore().LockOneGlobalTrans(expireIn)
	if global == nil {
		return nil
	}
	logger.Infof("cron job return a trans: %s", global.String())
	return &TransGlobal{TransGlobalStore: *global}
}

func handlePanic(perr *error) {
	if err := recover(); err != nil {
		logger.Errorf("----recovered panic %v\n%s", err, string(debug.Stack()))
		if perr != nil {
			*perr = fmt.Errorf("dtm panic: %v", err)
		}
	}
}

func sleepCronTime() {
	normal := time.Duration((float64(conf.TransCronInterval) - rand.Float64()) * float64(time.Second))
	interval := dtmimp.If(CronForwardDuration > 0, 1*time.Millisecond, normal).(time.Duration)
	logger.Debugf("sleeping for %v milli", interval/time.Microsecond)
	time.Sleep(interval)
}

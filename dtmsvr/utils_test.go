/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func TestUtils(t *testing.T) {
	db := dbGet()
	db.NoMust()
	err := dtmimp.CatchP(func() {
		checkAffected(db.DB)
	})
	assert.Error(t, err)

	CronExpiredTrans(1)
	sleepCronTime()
}

func TestCheckLocalHost(t *testing.T) {
	config.DisableLocalhost = 1
	err := dtmimp.CatchP(func() {
		checkLocalhost([]TransBranch{{URL: "http://localhost"}})
	})
	assert.Error(t, err)
	config.DisableLocalhost = 0
	err = dtmimp.CatchP(func() {
		checkLocalhost([]TransBranch{{URL: "http://localhost"}})
	})
	assert.Nil(t, err)
}

func TestSetNextCron(t *testing.T) {
	tg := TransGlobal{}
	tg.RetryInterval = 15
	tg.setNextCron(cronReset)
	assert.Equal(t, int64(15), tg.NextCronInterval)
	tg.RetryInterval = 0
	tg.setNextCron(cronReset)
	assert.Equal(t, config.RetryInterval, tg.NextCronInterval)
	tg.setNextCron(cronBackoff)
	assert.Equal(t, config.RetryInterval*2, tg.NextCronInterval)
}

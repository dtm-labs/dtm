/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"os"
	"testing"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func TestMain(m *testing.M) {
	common.MustLoadConfig()
	dtmcli.SetCurrentDBType(common.DtmConfig.DB["driver"])
	os.Setenv("DTM_DEBUG", "1")
	dtmsvr.TransProcessedTestChan = make(chan string, 1)
	dtmsvr.NowForwardDuration = 0 * time.Second
	dtmsvr.CronForwardDuration = 180 * time.Second
	common.DtmConfig.UpdateBranchSync = 1
	dtmsvr.PopulateDB(false)
	examples.PopulateDB(false)
	// 启动组件
	go dtmsvr.StartSvr()
	examples.GrpcStartup()
	app = examples.BaseAppStartup()

	m.Run()
}

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

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/workflow"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/go-resty/resty/v2"
)

func TestMain(m *testing.M) {
	config.MustLoadConfig("")
	logger.InitLog("debug")
	dtmsvr.TransProcessedTestChan = make(chan string, 1)
	dtmsvr.NowForwardDuration = 0 * time.Second
	dtmsvr.CronForwardDuration = 180 * time.Second
	conf.UpdateBranchSync = 1

	dtmgrpc.AddUnaryInterceptor(busi.SetGrpcHeaderForHeadersYes)
	dtmcli.GetRestyClient().OnBeforeRequest(busi.SetHTTPHeaderForHeadersYes)
	dtmcli.GetRestyClient().OnAfterResponse(func(c *resty.Client, resp *resty.Response) error { return nil })

	tenv := dtmimp.OrString(os.Getenv("TEST_STORE"), config.Redis)
	conf.Store.Host = "localhost"
	conf.Store.Driver = tenv
	if tenv == "boltdb" {
	} else if tenv == config.Mysql {
		conf.Store.Port = 3306
		conf.Store.User = "root"
		conf.Store.Password = ""
	} else if tenv == config.Postgres {
		conf.Store.Port = 5432
		conf.Store.User = "postgres"
		conf.Store.Password = "mysecretpassword"
	} else if tenv == config.Redis {
		conf.Store.User = ""
		conf.Store.Password = ""
		conf.Store.Port = 6379
	}
	conf.Store.Db = ""
	registry.WaitStoreUp()

	dtmsvr.PopulateDB(false)
	conf.Store.Db = "dtm" // after populateDB, set current db to dtm
	if tenv == "postgres" {
		busi.BusiConf = conf.Store.GetDBConf()
		dtmcli.SetCurrentDBType(tenv)
	}
	go dtmsvr.StartSvr()

	busi.PopulateDB(false)
	hsvr, gsvr := busi.Startup()
	// WorkflowStarup 1
	workflow.InitHTTP(dtmutil.DefaultHTTPServer, Busi+"/workflow/resume")
	workflow.InitGrpc(dtmutil.DefaultGrpcServer, busi.BusiGrpc, gsvr)
	go busi.RunGrpc(gsvr)
	go busi.RunHTTP(hsvr)
	r := m.Run()
	if r != 0 {
		os.Exit(r)
	}
	close(dtmsvr.TransProcessedTestChan)
	gid, more := <-dtmsvr.TransProcessedTestChan
	logger.FatalfIf(more, "extra gid: %s in test chan", gid)
	os.Exit(0)
}

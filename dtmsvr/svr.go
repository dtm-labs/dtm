/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtmdriver"
	"google.golang.org/grpc"
)

// StartSvr StartSvr
func StartSvr() {
	logger.Infof("start dtmsvr")
	setServerInfoMetrics()

	dtmcli.GetRestyClient().SetTimeout(time.Duration(conf.RequestTimeout) * time.Second)
	dtmgrpc.AddUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx2, cancel := context.WithTimeout(ctx, time.Duration(conf.RequestTimeout)*time.Second)
		defer cancel()
		return invoker(ctx2, method, req, reply, cc, opts...)
	})

	// start gin server
	app := dtmutil.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	logger.Infof("dtmsvr listen at: %d", conf.HTTPPort)
	go func() {
		err := app.Run(fmt.Sprintf(":%d", conf.HTTPPort))
		if err != nil {
			logger.Errorf("start server err: %v", err)
		}
	}()

	// start grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	logger.FatalIfError(err)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(grpcMetrics, dtmgimp.GrpcServerLog))
	dtmgpb.RegisterDtmServer(s, &dtmServer{})
	logger.Infof("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		logger.FatalIfError(err)
	}()

	for i := 0; i < int(conf.UpdateBranchAsyncGoroutineNum); i++ {
		go updateBranchAsync()
	}

	time.Sleep(100 * time.Millisecond)
	err = dtmdriver.Use(conf.MicroService.Driver)
	logger.FatalIfError(err)
	err = dtmdriver.GetDriver().RegisterGrpcService(conf.MicroService.Target, conf.MicroService.EndPoint)
	logger.FatalIfError(err)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	GetStore().PopulateData(skipDrop)
}

// UpdateBranchAsyncInterval interval to flush branch
var UpdateBranchAsyncInterval = 200 * time.Millisecond
var updateBranchAsyncChan chan branchStatus = make(chan branchStatus, 1000)

func updateBranchAsync() {
	flushBranchs := func() {
		defer dtmutil.RecoverPanic(nil)
		updates := []TransBranch{}
		started := time.Now()
		checkInterval := 20 * time.Millisecond
		for time.Since(started) < UpdateBranchAsyncInterval-checkInterval && len(updates) < 20 {
			select {
			case updateBranch := <-updateBranchAsyncChan:
				updates = append(updates, TransBranch{
					ModelBase:  dtmutil.ModelBase{ID: updateBranch.id},
					Gid:        updateBranch.gid,
					Status:     updateBranch.status,
					FinishTime: updateBranch.finishTime,
				})
			case <-time.After(checkInterval):
			}
		}
		for len(updates) > 0 {
			rowAffected, err := GetStore().UpdateBranches(updates, []string{"status", "finish_time", "update_time"})

			if err != nil {
				logger.Errorf("async update branch status error: %v", err)
				time.Sleep(1 * time.Second)
			} else {
				logger.Infof("flushed %d branch status to db. affected: %d", len(updates), rowAffected)
				updates = []TransBranch{}
			}
		}

	}
	for { // flush branches every 200ms
		flushBranchs()
	}
}

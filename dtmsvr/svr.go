/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/logger"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgpb"
	"github.com/yedf/dtmdriver"
	"google.golang.org/grpc"
)

// StartSvr StartSvr
func StartSvr() {
	logger.Debugf("start dtmsvr")
	app := common.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	logger.Debugf("dtmsvr listen at: %d", config.HttpPort)
	go app.Run(fmt.Sprintf(":%d", config.HttpPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	logger.FatalIfError(err)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc.UnaryServerInterceptor(grpcMetrics), grpc.UnaryServerInterceptor(dtmgimp.GrpcServerLog)),
		))
	dtmgpb.RegisterDtmServer(s, &dtmServer{})
	logger.Debugf("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		logger.FatalIfError(err)
	}()
	go updateBranchAsync()

	time.Sleep(100 * time.Millisecond)
	err = dtmdriver.Use(config.MicroService.Driver)
	logger.FatalIfError(err)
	err = dtmdriver.GetDriver().RegisterGrpcService(config.MicroService.Target, config.MicroService.EndPoint)
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
	for { // flush branches every second
		defer common.RecoverPanic(nil)
		updates := []TransBranch{}
		started := time.Now()
		checkInterval := 20 * time.Millisecond
		for time.Since(started) < UpdateBranchAsyncInterval-checkInterval && len(updates) < 20 {
			select {
			case updateBranch := <-updateBranchAsyncChan:
				updates = append(updates, TransBranch{
					ModelBase:  common.ModelBase{ID: updateBranch.id},
					Status:     updateBranch.status,
					FinishTime: updateBranch.finishTime,
				})
			case <-time.After(checkInterval):
			}
		}
		for len(updates) > 0 {
			dbr := GetStore().UpdateBranchesSql(updates, []string{"status", "finish_time", "update_time"})

			logger.Debugf("flushed %d branch status to db. affected: %d", len(updates), dbr.RowsAffected)
			if dbr.Error != nil {
				logger.Errorf("async update branch status error: %v", dbr.Error)
				time.Sleep(1 * time.Second)
			} else {
				updates = []TransBranch{}
			}
		}
	}
}

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
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"github.com/yedf/dtm/examples"
	"github.com/yedf/dtmdriver"
	"google.golang.org/grpc"
	"gorm.io/gorm/clause"

	// _ "github.com/ychensha/dtmdriver-polaris"
	_ "github.com/yedf/dtmdriver-gozero"
	_ "github.com/yedf/dtmdriver-protocol1"
)

// StartSvr StartSvr
func StartSvr() {
	dtmimp.Logf("start dtmsvr")
	app := common.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	dtmimp.Logf("dtmsvr listen at: %d", common.DtmHttpPort)
	go app.Run(fmt.Sprintf(":%d", common.DtmHttpPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", common.DtmGrpcPort))
	dtmimp.FatalIfError(err)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc.UnaryServerInterceptor(grpcMetrics), grpc.UnaryServerInterceptor(dtmgimp.GrpcServerLog)),
		))
	dtmgimp.RegisterDtmServer(s, &dtmServer{})
	dtmimp.Logf("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		dtmimp.FatalIfError(err)
	}()
	go updateBranchAsync()

	// prometheus exporter
	dtmimp.Logf("prometheus exporter listen at: %d", common.DtmMetricsPort)
	prometheusHTTPRun(fmt.Sprintf("%d", common.DtmMetricsPort))
	time.Sleep(100 * time.Millisecond)
	err = dtmdriver.Use(config.MicroService.Driver)
	dtmimp.FatalIfError(err)
	err = dtmdriver.GetDriver().RegisterGrpcService(config.MicroService.Target, config.MicroService.EndPoint)
	dtmimp.FatalIfError(err)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCallerCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

// UpdateBranchAsyncInterval interval to flush branch
var UpdateBranchAsyncInterval = 200 * time.Millisecond
var updateBranchAsyncChan chan branchStatus = make(chan branchStatus, 1000)

func updateBranchAsync() {
	for { // flush branches every second
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
			dbr := dbGet().Clauses(clause.OnConflict{
				OnConstraint: "trans_branch_op_pkey",
				DoUpdates:    clause.AssignmentColumns([]string{"status", "finish_time", "update_time"}),
			}).Create(updates)
			dtmimp.Logf("flushed %d branch status to db. affected: %d", len(updates), dbr.RowsAffected)
			if dbr.Error != nil {
				dtmimp.LogRedf("async update branch status error: %v", dbr.Error)
				time.Sleep(1 * time.Second)
			} else {
				updates = []TransBranch{}
			}
		}
	}
}

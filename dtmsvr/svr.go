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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtmdriver"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// StartSvr StartSvr
func StartSvr() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	logger.Infof("start dtmsvr")
	app := dtmutil.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", conf.HttpPort),
		Handler: app,
	}
	logger.Infof("dtmsvr http listening at: %d", conf.HttpPort)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("dtmsvr run error: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	logger.FatalIfError(err)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpcMetrics,
		dtmgimp.GrpcServerLog,
	)))
	dtmgpb.RegisterDtmServer(grpcServer, &dtmServer{})
	logger.Infof("dtmsvr grpc listening at: %d", conf.GrpcPort)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.Errorf("dtmsvr run error: %v", err)
		}
	}()
	go updateBranchAsync()

	time.Sleep(100 * time.Millisecond)
	err = dtmdriver.Use(conf.MicroService.Driver)
	logger.FatalIfError(err)
	err = dtmdriver.GetDriver().RegisterGrpcService(conf.MicroService.Target, conf.MicroService.EndPoint)
	logger.FatalIfError(err)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	grpcServer.GracefulStop()
	logger.Infof("dtmsvr grpc stopped")
	err = httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf("dtmsvr http shutdown error: %v", err)
	}
	logger.Infof("dtmsvr http stopped")
	<-ctx.Done()
	close(stop)
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
}

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

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/gin-gonic/gin"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// StartSvr StartSvr
func StartSvr() *gin.Engine {
	logger.Infof("start dtmsvr")
	setServerInfoMetrics()

	dtmcli.GetRestyClient().SetTimeout(time.Duration(conf.RequestTimeout) * time.Second)
	dtmgrpc.AddUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		timeout := conf.RequestTimeout
		if v := dtmgimp.RequestTimeoutFromContext(ctx); v != 0 {
			timeout = v
		}
		ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		return invoker(ctx2, method, req, reply, cc, opts...)
	})

	// start gin server
	app := dtmutil.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	addJrpcRouter(app)
	logger.Infof("dtmsvr http listen at: %d", conf.HTTPPort)
	go func() {
		err := app.Run(fmt.Sprintf(":%d", conf.HTTPPort))
		if err != nil {
			logger.Errorf("start server err: %v", err)
		}
	}()

	// start grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	logger.FatalIfError(err)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(grpcRecover, grpcMetrics, dtmgimp.GrpcServerLog))
	dtmgpb.RegisterDtmServer(s, &dtmServer{})
	reflection.Register(s)
	logger.Infof("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		logger.FatalIfError(err)
	}()

	for i := 0; i < int(conf.UpdateBranchAsyncGoroutineNum); i++ {
		go updateBranchAsync()
	}
	updateTopicsMap()
	go CronUpdateTopicsMap()

	time.Sleep(100 * time.Millisecond)
	err = dtmdriver.Use(conf.MicroService.Driver)
	logger.FatalIfError(err)
	logger.Infof("RegisterService: %s", conf.MicroService.Driver)
	err = dtmdriver.GetDriver().RegisterService(conf.MicroService.Target, conf.MicroService.EndPoint)
	logger.FatalIfError(err)
	return app
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	GetStore().PopulateData(skipDrop)
}

// UpdateBranchAsyncInterval interval to flush branch
var UpdateBranchAsyncInterval = 200 * time.Millisecond
var updateBranchAsyncChan = make(chan branchStatus, 1000)

func updateBranchAsync() {
	flushBranchs := func() {
		defer dtmutil.RecoverPanic(nil)
		updates := []TransBranch{}
		exists := map[string]bool{}
		started := time.Now()
		checkInterval := 20 * time.Millisecond
		for time.Since(started) < UpdateBranchAsyncInterval-checkInterval && len(updates) < 20 {
			select {
			case updateBranch := <-updateBranchAsyncChan:
				k := updateBranch.gid + updateBranch.branchID + "-" + updateBranch.op
				if !exists[k] { // postgres does not allow
					exists[k] = true
					updates = append(updates, TransBranch{
						Gid:        updateBranch.gid,
						BranchID:   updateBranch.branchID,
						Op:         updateBranch.op,
						Status:     updateBranch.status,
						FinishTime: updateBranch.finishTime,
					})
				}
			case <-time.After(checkInterval):
			}
		}
		for i := 0; i < 3 && len(updates) > 0; i++ {
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

func grpcRecover(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, rerr error) {
	defer func() {
		if x := recover(); x != nil {
			rerr = status.Errorf(codes.Internal, "%v", x)
			logger.Errorf("dtm server panic: %v", x)
		}
	}()
	res, rerr = handler(ctx, req)
	return
}

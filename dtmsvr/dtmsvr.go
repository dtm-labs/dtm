package dtmsvr

import (
	"fmt"
	"net"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"
	"gorm.io/gorm/clause"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"

	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080
var dtmsvrGrpcPort = 58080
var metricsPort = 8889

// StartSvr StartSvr
func StartSvr() {
	dtmcli.Logf("start dtmsvr")
	app := common.GetGinApp()
	app = httpMetrics(app)
	addRoute(app)
	dtmcli.Logf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dtmsvrGrpcPort))
	dtmcli.FatalIfError(err)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc.UnaryServerInterceptor(grpcMetrics), grpc.UnaryServerInterceptor(dtmgrpc.GrpcServerLog)),
		))
	dtmgrpc.RegisterDtmServer(s, &dtmServer{})
	dtmcli.Logf("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		dtmcli.FatalIfError(err)
	}()
	go updateBranchAsync()

	// prometheus exporter
	dtmcli.Logf("prometheus exporter listen at: %d", metricsPort)
	prometheusHTTPRun(fmt.Sprintf("%d", metricsPort))
	time.Sleep(100 * time.Millisecond)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCallerCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

// UpdateBranchAsyncInterval unit millisecond
var UpdateBranchAsyncInterval time.Duration = 1000
var updateBranchAsyncChan chan branchStatus = make(chan branchStatus, 1000)

func updateBranchAsync() {
	for { // flush branches every second
		updates := []TransBranch{}
		started := time.Now()
		for time.Since(started) < UpdateBranchAsyncInterval*time.Millisecond {
			select {
			case updateBranch := <-updateBranchAsyncChan:
				updates = append(updates, TransBranch{
					ModelBase:  common.ModelBase{ID: updateBranch.id},
					Status:     updateBranch.status,
					FinishTime: updateBranch.finish_time,
				})
			case <-time.After(50 * time.Millisecond):
			}
		}
		for len(updates) > 0 {
			dbr := dbGet().Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"status", "finish_time"}),
			}).Create(updates)
			dtmcli.Logf("flushed %d branch status to db. affected: %d", len(updates), dbr.RowsAffected)
			if dbr.Error != nil {
				dtmcli.LogRedf("async update branch status error: %v", dbr.Error)
				time.Sleep(1 * time.Second)
			} else {
				updates = []TransBranch{}
			}
		}
	}
}

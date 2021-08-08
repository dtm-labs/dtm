package dtmsvr

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
	"google.golang.org/grpc"
)

var dtmsvrPort = 8080
var dtmsvrGrpcPort = 50051

// StartSvr StartSvr
func StartSvr() {
	dtmcli.Logf("start dtmsvr")
	app := common.GetGinApp()
	addRoute(app)
	dtmcli.Logf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dtmsvrGrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	dtmcli.RegisterDtmServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCurrentCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

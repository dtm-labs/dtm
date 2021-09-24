package dtmsvr

import (
	"fmt"
	"net"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmgrpc"

	"google.golang.org/grpc"

	"github.com/yedf/dtm/examples"
)

var dtmsvrPort = 8080
var dtmsvrGrpcPort = 58080

// StartSvr StartSvr
func StartSvr() {
	dtmcli.Logf("start dtmsvr")
	app := common.GetGinApp()
	addRoute(app)
	dtmcli.Logf("dtmsvr listen at: %d", dtmsvrPort)
	go app.Run(fmt.Sprintf(":%d", dtmsvrPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dtmsvrGrpcPort))
	dtmcli.FatalIfError(err)
	s := grpc.NewServer(grpc.UnaryInterceptor(dtmgrpc.GrpcServerLog))
	dtmgrpc.RegisterDtmServer(s, &dtmServer{})
	dtmcli.Logf("grpc listening at %v", lis.Addr())
	go func() {
		err := s.Serve(lis)
		dtmcli.FatalIfError(err)
	}()

	time.Sleep(100 * time.Millisecond)
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCallerCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

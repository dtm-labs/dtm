package busi

import (
	"github.com/dtm-labs/dtm/dtmgrpc/workflow"
	"github.com/dtm-labs/dtm/dtmutil"
	"google.golang.org/grpc"
)

func WorkflowStarup(server *grpc.Server) {
	workflow.InitHttp(dtmServer, Busi+"/workflow/resume")
	workflow.InitGrpc(dtmutil.DefaultGrpcServer, BusiGrpc, server)
}

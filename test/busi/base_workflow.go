package busi

import (
	"github.com/dtm-labs/dtm/dtmgrpc/workflow"
	"github.com/dtm-labs/dtm/dtmutil"
	"google.golang.org/grpc"
)

// WorkflowStarup 1
func WorkflowStarup(server *grpc.Server) {
	workflow.InitHTTP(dtmServer, Busi+"/workflow/resume")
	workflow.InitGrpc(dtmutil.DefaultGrpcServer, BusiGrpc, server)
}

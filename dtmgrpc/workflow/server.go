package workflow

import (
	"context"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/workflow/wfpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type workflowServer struct {
	wfpb.UnimplementedWorkflowServer
}

func (s *workflowServer) Execute(ctx context.Context, wd *wfpb.WorkflowData) (*emptypb.Empty, error) {
	if defaultFac.protocol != dtmimp.ProtocolGRPC {
		return nil, status.Errorf(codes.Internal, "workflow server not inited. please call workflow.InitGrpc first")
	}
	tb := dtmgimp.TransBaseFromGrpc(ctx)
	err := defaultFac.execute(tb.Op, tb.Gid, wd.Data)
	return &emptypb.Empty{}, dtmgrpc.DtmError2GrpcError(err)
}

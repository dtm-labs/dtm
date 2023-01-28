package test

import (
	"context"
	"testing"

	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWorkflowInterceptorOutsideSaga(t *testing.T) {
	called := false
	workflow.Interceptor(context.TODO(), "method", nil, nil, &grpc.ClientConn{}, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		called = true
		return nil
	})
	assert.True(t, called)
}

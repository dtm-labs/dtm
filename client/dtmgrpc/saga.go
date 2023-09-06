/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"context"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"google.golang.org/protobuf/proto"
)

// SagaGrpc struct of saga
type SagaGrpc struct {
	dtmcli.Saga
}

// NewSagaGrpc create a saga
func NewSagaGrpc(server string, gid string, opts ...TransBaseOption) *SagaGrpc {
	sg := &SagaGrpc{Saga: *dtmcli.NewSaga(server, gid)}

	for _, opt := range opts {
		opt(&sg.TransBase)
	}

	return sg
}

// NewSagaGrpcWithContext create a saga with context
func NewSagaGrpcWithContext(ctx context.Context, server string, gid string, opts ...TransBaseOption) *SagaGrpc {
	sg := &SagaGrpc{Saga: *dtmcli.NewSagaWithContext(ctx, server, gid)}

	for _, opt := range opts {
		opt(&sg.TransBase)
	}

	return sg
}

// Add add a saga step
func (s *SagaGrpc) Add(action string, compensate string, payload proto.Message) *SagaGrpc {
	s.Steps = append(s.Steps, map[string]string{"action": action, "compensate": compensate})
	s.BinPayloads = append(s.BinPayloads, dtmgimp.MustProtoMarshal(payload))
	return s
}

// AddBranchOrder specify that branch should be after preBranches. branch should is larger than all the element in preBranches
func (s *SagaGrpc) AddBranchOrder(branch int, preBranches []int) *SagaGrpc {
	s.Saga.AddBranchOrder(branch, preBranches)
	return s
}

// EnableConcurrent enable the concurrent exec of sub trans
func (s *SagaGrpc) EnableConcurrent() *SagaGrpc {
	s.Saga.SetConcurrent()
	return s
}

// Submit submit the saga trans
func (s *SagaGrpc) Submit() error {
	s.Saga.BuildCustomOptions()
	return dtmgimp.DtmGrpcCall(&s.Saga.TransBase, "Submit")
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/protobuf/proto"
)

// SagaGrpc struct of saga
type SagaGrpc struct {
	dtmcli.Saga
}

// NewSagaGrpc create a saga
func NewSagaGrpc(server string, gid string) *SagaGrpc {
	return &SagaGrpc{Saga: *dtmcli.NewSaga(server, gid)}
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

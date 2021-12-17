/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"github.com/yedf/dtm/dtmsvr/storage"
)

// TransGlobal global transaction
type TransGlobal struct {
	storage.TransGlobalStore
	lastTouched      time.Time // record the start time of process
	updateBranchSync bool
}

// TransBranch branch transaction
type TransBranch = storage.TransBranchStore

type transProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(branches []TransBranch) error
}

type processorCreator func(*TransGlobal) transProcessor

var processorFac = map[string]processorCreator{}

func registorProcessorCreator(transType string, creator processorCreator) {
	processorFac[transType] = creator
}

func (t *TransGlobal) getProcessor() transProcessor {
	return processorFac[t.TransType](t)
}

type cronType int

const (
	cronBackoff cronType = iota
	cronReset
	cronKeep
)

// TransFromContext TransFromContext
func TransFromContext(c *gin.Context) *TransGlobal {
	b, err := c.GetRawData()
	e2p(err)
	m := TransGlobal{}
	dtmimp.MustUnmarshal(b, &m)
	dtmimp.Logf("creating trans in prepare")
	// Payloads will be store in BinPayloads, Payloads is only used to Unmarshal
	for _, p := range m.Payloads {
		m.BinPayloads = append(m.BinPayloads, []byte(p))
	}
	for _, d := range m.Steps {
		if d["data"] != "" {
			m.BinPayloads = append(m.BinPayloads, []byte(d["data"]))
		}
	}
	m.Protocol = "http"
	return &m
}

// TransFromDtmRequest TransFromContext
func TransFromDtmRequest(c *dtmgimp.DtmRequest) *TransGlobal {
	o := &dtmgimp.DtmTransOptions{}
	if c.TransOptions != nil {
		o = c.TransOptions
	}
	r := TransGlobal{TransGlobalStore: storage.TransGlobalStore{
		Gid:           c.Gid,
		TransType:     c.TransType,
		QueryPrepared: c.QueryPrepared,
		Protocol:      "grpc",
		BinPayloads:   c.BinPayloads,
		TransOptions: dtmcli.TransOptions{
			WaitResult:    o.WaitResult,
			TimeoutToFail: o.TimeoutToFail,
			RetryInterval: o.RetryInterval,
		},
	}}
	if c.Steps != "" {
		dtmimp.MustUnmarshalString(c.Steps, &r.Steps)
	}
	return &r
}

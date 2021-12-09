/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"gorm.io/gorm"
)

var errUniqueConflict = errors.New("unique key conflict error")

// TransGlobal global transaction
type TransGlobal struct {
	common.ModelBase
	Gid              string              `json:"gid"`
	TransType        string              `json:"trans_type"`
	Steps            []map[string]string `json:"steps" gorm:"-"`
	Payloads         []string            `json:"payloads" gorm:"-"`
	BinPayloads      [][]byte            `json:"-" gorm:"-"`
	Status           string              `json:"status"`
	QueryPrepared    string              `json:"query_prepared"`
	Protocol         string              `json:"protocol"`
	CommitTime       *time.Time
	FinishTime       *time.Time
	RollbackTime     *time.Time
	Options          string
	CustomData       string `json:"custom_data"`
	NextCronInterval int64
	NextCronTime     *time.Time
	dtmcli.TransOptions
	lastTouched      time.Time // record the start time of process
	updateBranchSync bool
}

// TableName TableName
func (*TransGlobal) TableName() string {
	return "dtm.trans_global"
}

// TransBranch branch transaction
type TransBranch struct {
	common.ModelBase
	Gid          string
	URL          string `json:"url"`
	BinData      []byte
	BranchID     string `json:"branch_id"`
	Op           string
	Status       string
	FinishTime   *time.Time
	RollbackTime *time.Time
}

// TableName TableName
func (*TransBranch) TableName() string {
	return "dtm.trans_branch_op"
}

type transProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(db *common.DB, branches []TransBranch) error
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
	r := TransGlobal{
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
	}
	if c.Steps != "" {
		dtmimp.MustUnmarshalString(c.Steps, &r.Steps)
	}
	return &r
}

func checkAffected(db1 *gorm.DB) {
	if db1.RowsAffected == 0 {
		panic(fmt.Errorf("rows affected 0, please check for abnormal trans"))
	}
}

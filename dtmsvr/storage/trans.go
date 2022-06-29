/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package storage

import (
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
)

// TransGlobalExt defines Header info
type TransGlobalExt struct {
	Headers map[string]string `json:"headers,omitempty" gorm:"-"`
}

// TransGlobalStore defines GlobalStore storage info
type TransGlobalStore struct {
	dtmutil.ModelBase
	Gid              string              `json:"gid,omitempty"`
	TransType        string              `json:"trans_type,omitempty"`
	Steps            []map[string]string `json:"steps,omitempty" gorm:"-"`
	Payloads         []string            `json:"payloads,omitempty" gorm:"-"`
	BinPayloads      [][]byte            `json:"-" gorm:"-"`
	Status           string              `json:"status,omitempty"`
	QueryPrepared    string              `json:"query_prepared,omitempty"`
	Protocol         string              `json:"protocol,omitempty"`
	FinishTime       *time.Time          `json:"finish_time,omitempty"`
	RollbackTime     *time.Time          `json:"rollback_time,omitempty"`
	RollbackReason   string              `json:"rollback_reason,omitempty"`
	Options          string              `json:"options,omitempty"`
	CustomData       string              `json:"custom_data,omitempty"`
	NextCronInterval int64               `json:"next_cron_interval,omitempty"`
	NextCronTime     *time.Time          `json:"next_cron_time,omitempty"`
	Owner            string              `json:"owner,omitempty"`
	Ext              TransGlobalExt      `json:"-" gorm:"-"`
	ExtData          string              `json:"ext_data,omitempty"` // storage of ext. a db field to store many values. like Options
	dtmcli.TransOptions
}

// TableName TableName
func (g *TransGlobalStore) TableName() string {
	return "trans_global"
}

func (g *TransGlobalStore) String() string {
	return dtmimp.MustMarshalString(g)
}

// IsFinished return true if status == "failed" || status == "succeed"
func (g *TransGlobalStore) IsFinished() bool {
	return g.Status == dtmcli.StatusFailed || g.Status == dtmcli.StatusSucceed
}

// TransBranchStore branch transaction
type TransBranchStore struct {
	dtmutil.ModelBase
	Gid          string     `json:"gid,omitempty"`
	URL          string     `json:"url,omitempty"`
	BinData      []byte     `json:"bin_data,omitempty"`
	BranchID     string     `json:"branch_id,omitempty"`
	Op           string     `json:"op,omitempty"`
	Status       string     `json:"status,omitempty"`
	FinishTime   *time.Time `json:"finish_time,omitempty"`
	RollbackTime *time.Time `json:"rollback_time,omitempty"`
	Error        error      `json:"-" gorm:"-"`
}

// TableName TableName
func (b *TransBranchStore) TableName() string {
	return "trans_branch_op"
}

func (b *TransBranchStore) String() string {
	return dtmimp.MustMarshalString(*b)
}

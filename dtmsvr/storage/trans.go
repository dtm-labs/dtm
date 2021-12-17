package storage

import (
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

type TransGlobalStore struct {
	common.ModelBase
	Gid              string              `json:"gid,omitempty"`
	TransType        string              `json:"trans_type,omitempty"`
	Steps            []map[string]string `json:"steps,omitempty" gorm:"-"`
	Payloads         []string            `json:"payloads,omitempty" gorm:"-"`
	BinPayloads      [][]byte            `json:"-" gorm:"-"`
	Status           string              `json:"status,omitempty"`
	QueryPrepared    string              `json:"query_prepared,omitempty"`
	Protocol         string              `json:"protocol,omitempty"`
	CommitTime       *time.Time          `json:"commit_time,omitempty"`
	FinishTime       *time.Time          `json:"finish_time,omitempty"`
	RollbackTime     *time.Time          `json:"rollback_time,omitempty"`
	Options          string              `json:"options,omitempty"`
	CustomData       string              `json:"custom_data,omitempty"`
	NextCronInterval int64               `json:"next_cron_interval,omitempty"`
	NextCronTime     *time.Time          `json:"next_cron_time,omitempty"`
	Owner            string              `json:"owner,omitempty"`
	dtmcli.TransOptions
}

// TableName TableName
func (*TransGlobalStore) TableName() string {
	return "dtm.trans_global"
}

// TransBranchStore branch transaction
type TransBranchStore struct {
	common.ModelBase
	Gid          string `json:"gid,omitempty"`
	URL          string `json:"url,omitempty"`
	BinData      []byte
	BranchID     string     `json:"branch_id,omitempty"`
	Op           string     `json:"op,omitempty"`
	Status       string     `json:"status,omitempty"`
	FinishTime   *time.Time `json:"finish_time,omitempty"`
	RollbackTime *time.Time `json:"rollback_time,omitempty"`
}

// TableName TableName
func (*TransBranchStore) TableName() string {
	return "dtm.trans_branch_op"
}

package dtmsvr

import (
	"time"

	"github.com/yedf/dtm/common"
)

type M = map[string]interface{}

type TransGlobalModel struct {
	common.ModelBase
	Gid           string `json:"gid"`
	TransType     string `json:"trans_type"`
	Data          string `json:"data"`
	Status        string `json:"status"`
	QueryPrepared string `json:"query_prepared"`
	CommitTime    *time.Time
	FinishTime    *time.Time
	RollbackTime  *time.Time
}

func (*TransGlobalModel) TableName() string {
	return "trans_global"
}

type TransBranchModel struct {
	common.ModelBase
	Gid          string
	Url          string
	Data         string
	Branch       string
	BranchType   string
	Status       string
	FinishTime   *time.Time
	RollbackTime *time.Time
}

func (*TransBranchModel) TableName() string {
	return "trans_branch"
}

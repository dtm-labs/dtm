package examples

import "github.com/yedf/dtm/common"

var e2p = common.E2P

type UserAccount struct {
	common.ModelBase
	UserId  int
	Balance string
}

func (u *UserAccount) TableName() string { return "user_account" }

package examples

import "github.com/yedf/dtm/common"

type UserAccount struct {
	common.ModelBase
	UserId  int
	Balance string
}

func (u *UserAccount) TableName() string { return "user_account" }

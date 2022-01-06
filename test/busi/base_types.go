/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

// BusiConf defines busi conf
var BusiConf = dtmcli.DBConf{
	Driver: "mysql",
	Host:   "localhost",
	Port:   3306,
	User:   "root",
}

// UserAccount defines user account info
type UserAccount struct {
	UserID         int
	Balance        string
	TradingBalance string
}

// TableName returns UserAccount's table name
func (*UserAccount) TableName() string {
	return "dtm_busi.user_account"
}

// GetUserAccountByUID finds UserAccount by uid
func GetUserAccountByUID(uid int) *UserAccount {
	ua := UserAccount{}
	dbr := dbGet().Must().Model(&ua).Where("user_id=?", uid).First(&ua)
	dtmimp.E2P(dbr.Error)
	return &ua
}

// IsEqual checks ua1 and ua2 is equal
func IsEqual(ua1, ua2 *UserAccount) bool {
	return ua1.UserID == ua2.UserID && ua1.Balance == ua2.Balance && ua1.TradingBalance == ua2.TradingBalance
}

// TransReq transaction request payload
type TransReq struct {
	Amount         int    `json:"amount"`
	TransInResult  string `json:"trans_in_result"`
	TransOutResult string `json:"trans_out_Result"`
}

func (t *TransReq) String() string {
	return fmt.Sprintf("amount: %d transIn: %s transOut: %s", t.Amount, t.TransInResult, t.TransOutResult)
}

// GenTransReq 1
func GenTransReq(amount int, outFailed bool, inFailed bool) *TransReq {
	return &TransReq{
		Amount:         amount,
		TransOutResult: dtmimp.If(outFailed, dtmcli.ResultFailure, "").(string),
		TransInResult:  dtmimp.If(inFailed, dtmcli.ResultFailure, "").(string),
	}
}

// GenBusiReq 1
func GenBusiReq(amount int, outFailed bool, inFailed bool) *BusiReq {
	return &BusiReq{
		Amount:         int64(amount),
		TransOutResult: dtmimp.If(outFailed, dtmcli.ResultFailure, "").(string),
		TransInResult:  dtmimp.If(inFailed, dtmcli.ResultFailure, "").(string),
	}
}

func reqFrom(c *gin.Context) *TransReq {
	v, ok := c.Get("trans_req")
	if !ok {
		req := TransReq{}
		err := c.BindJSON(&req)
		logger.FatalIfError(err)
		c.Set("trans_req", &req)
		v = &req
	}
	return v.(*TransReq)
}

func infoFromContext(c *gin.Context) *dtmcli.BranchBarrier {
	info := dtmcli.BranchBarrier{
		TransType: c.Query("trans_type"),
		Gid:       c.Query("gid"),
		BranchID:  c.Query("branch_id"),
		Op:        c.Query("op"),
	}
	return &info
}

// AutoEmptyString auto reset to empty when used once
type AutoEmptyString struct {
	value string
}

// SetOnce set a value once
func (s *AutoEmptyString) SetOnce(v string) {
	s.value = v
}

// Fetch fetch the stored value, then reset the value to empty
func (s *AutoEmptyString) Fetch() string {
	v := s.value
	s.value = ""
	return v
}

type mainSwitchType struct {
	TransInResult         AutoEmptyString
	TransOutResult        AutoEmptyString
	TransInConfirmResult  AutoEmptyString
	TransOutConfirmResult AutoEmptyString
	TransInRevertResult   AutoEmptyString
	TransOutRevertResult  AutoEmptyString
	CanSubmitResult       AutoEmptyString
	NextResult            AutoEmptyString
}

// MainSwitch controls busi success or fail
var MainSwitch mainSwitchType

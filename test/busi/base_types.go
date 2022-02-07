/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"context"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// StoreHost examples will connect to dtm.pub; tests will connect to localhost
var StoreHost = "localhost"

// BusiConf 1
var BusiConf = dtmcli.DBConf{
	Driver: "mysql",
	Host:   StoreHost,
	Port:   3306,
	User:   "root",
}

// UserAccount 1
type UserAccount struct {
	UserID         int
	Balance        string
	TradingBalance string
}

// TableName 1
func (*UserAccount) TableName() string {
	return "dtm_busi.user_account"
}

// GetBalanceByUID 1
func GetBalanceByUID(uid int, store string) int {
	if store == "redis" {
		rd := RedisGet()
		accA, err := rd.Get(rd.Context(), GetRedisAccountKey(uid)).Result()
		dtmimp.E2P(err)
		return dtmimp.MustAtoi(accA)
	} else if store == "mongo" {
		mg := MongoGet()
		account := mg.Database("dtm_busi").Collection("user_account")
		var result bson.M
		err := account.FindOne(context.Background(), bson.D{{Key: "user_id", Value: uid}}).Decode(&result)
		dtmimp.E2P(err)
		return int(result["balance"].(int32))
	}
	ua := UserAccount{}
	_ = dbGet().Must().Model(&ua).Where("user_id=?", uid).First(&ua)
	return dtmimp.MustAtoi(ua.Balance[:len(ua.Balance)-3])
}

// TransReq transaction request payload
type TransReq struct {
	Amount         int    `json:"amount"`
	TransInResult  string `json:"trans_in_result"`
	TransOutResult string `json:"trans_out_Result"`
	Store          string `json:"store"` // default mysql, value can be mysql|redis
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
	QueryPreparedResult   AutoEmptyString
	NextResult            AutoEmptyString
}

// MainSwitch controls busi success or fail
var MainSwitch mainSwitchType

// GetRedisAccountKey return redis key for uid
func GetRedisAccountKey(uid int) string {
	return fmt.Sprintf("{a}-redis-account-key-%d", uid)
}

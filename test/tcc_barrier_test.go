/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestTccBarrierNormal(t *testing.T) {
	req := busi.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccBarrierRollback(t *testing.T) {
	req := busi.GenTransReq(30, false, true)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestTccBarrierDisorderMysql(t *testing.T) {
	runTestTccBarrierDisorder(t, "mysql")
}

func TestTccBarrierDisorderMongo(t *testing.T) {
	runTestTccBarrierDisorder(t, "mongo")
}

func TestTccBarrierDisorderRedis(t *testing.T) {
	busi.SetRedisBothAccount(200, 200)
	runTestTccBarrierDisorder(t, "redis")
}

func runTestTccBarrierDisorder(t *testing.T, store string) {
	before := getBeforeBalances(store)
	cancelFinishedChan := make(chan string, 2)
	cancelCanReturnChan := make(chan string, 2)
	gid := dtmimp.GetFuncName() + store
	cronFinished := make(chan string, 2)
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		body := &busi.TransReq{Amount: 30, Store: store}
		tryURL := Busi + "/TccBTransOutTry"
		confirmURL := Busi + "/TccBTransOutConfirm"
		cancelURL := Busi + "/TccBSleepCancel"
		// 请参见子事务屏障里的时序图，这里为了模拟该时序图，手动拆解了callbranch
		branchID := tcc.NewSubBranchID()
		busi.SetSleepCancelHandler(func(c *gin.Context) interface{} {
			res := busi.TccBarrierTransOutCancel(c)
			logger.Debugf("disorderHandler before cancel finish write")
			cancelFinishedChan <- "1"
			logger.Debugf("disorderHandler before cancel return read")
			<-cancelCanReturnChan
			logger.Debugf("disorderHandler after cancel return read")
			return res
		})
		// 注册子事务
		resp, err := dtmimp.RestyClient.R().
			SetBody(map[string]interface{}{
				"gid":                tcc.Gid,
				"branch_id":          branchID,
				"trans_type":         "tcc",
				"status":             StatusPrepared,
				"data":               string(dtmimp.MustMarshal(body)),
				dtmcli.BranchConfirm: confirmURL,
				dtmcli.BranchCancel:  cancelURL,
			}).Post(fmt.Sprintf("%s/%s", tcc.Dtm, "registerBranch"))
		assert.Nil(t, err)
		assert.Contains(t, resp.String(), dtmcli.ResultSuccess)

		logger.Debugf("cron to timeout and then call cancelled twice")
		cron := func() {
			cronTransOnceForwardNow(t, gid, 300)
			logger.Debugf("cronFinished write")
			cronFinished <- "1"
			logger.Debugf("cronFinished after write")
		}
		go cron()
		<-cancelFinishedChan
		go cron()
		<-cancelFinishedChan
		cancelCanReturnChan <- "1"
		cancelCanReturnChan <- "1"
		logger.Debugf("after cancelCanRetrun 2 write")
		// after cancel then run try
		r, _ := dtmimp.RestyClient.R().
			SetBody(body).
			SetQueryParams(map[string]string{
				"dtm":        tcc.Dtm,
				"gid":        tcc.Gid,
				"branch_id":  branchID,
				"trans_type": "tcc",
				"op":         dtmcli.BranchTry,
			}).
			Post(tryURL)
		assert.True(t, strings.Contains(r.String(), dtmcli.ResultSuccess)) // 这个是悬挂操作，为了简单起见，依旧让他返回成功
		logger.Debugf("cronFinished read")
		<-cronFinished
		<-cronFinished
		logger.Debugf("cronFinished after read")

		return nil, fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assertSameBalance(t, before, store)
}

func TestTccBarrierPanic(t *testing.T) {
	bb := &dtmcli.BranchBarrier{TransType: "saga", Gid: "gid1", BranchID: "bid1", Op: "action", BarrierID: 1}
	var err error
	func() {
		defer dtmimp.P2E(&err)
		tx, _ := dbGet().ToSQLDB().BeginTx(context.Background(), &sql.TxOptions{})
		bb.Call(tx, func(tx *sql.Tx) error {
			panic(fmt.Errorf("an error"))
		})
	}()
	assert.Error(t, err, fmt.Errorf("an error"))
}

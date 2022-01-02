/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package svr

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid"
)

// launch command：go run app/main.go qs

// service address of the transcation
const benchAPI = "/api/busi_bench"
const total = 200000

var benchPort = dtmimp.If(os.Getenv("BENCH_PORT") == "", "8083", os.Getenv("BENCH_PORT")).(string)
var benchBusi = fmt.Sprintf("http://localhost:%s%s", benchPort, benchAPI)

func pdbGet() *sql.DB {
	db, err := dtmimp.PooledDB(busi.BusiConf)
	logger.FatalIfError(err)
	return db
}

func txGet() *sql.Tx {
	db := pdbGet()
	tx, err := db.Begin()
	logger.FatalIfError(err)
	return tx
}

func reloadData() {
	time.Sleep(dtmsvr.UpdateBranchAsyncInterval * 2)
	began := time.Now()
	db := pdbGet()
	tables := []string{"dtm_busi.user_account", "dtm_busi.user_account_log", "dtm.trans_global", "dtm.trans_branch_op", "dtm_barrier.barrier"}
	for _, t := range tables {
		_, err := dtmimp.DBExec(db, fmt.Sprintf("truncate %s", t))
		logger.FatalIfError(err)
	}
	s := "insert ignore into dtm_busi.user_account(user_id, balance) values "
	ss := []string{}
	for i := 1; i <= total; i++ {
		ss = append(ss, fmt.Sprintf("(%d, 1000000)", i))
	}
	_, err := dtmimp.DBExec(db, s+strings.Join(ss, ","))
	logger.FatalIfError(err)
	logger.Debugf("%d users inserted. used: %dms", total, time.Since(began).Milliseconds())
}

var uidCounter int32 = 0
var mode string = ""
var sqls int = 1

func PrepareBenchDB() {
	db := pdbGet()
	_, err := dtmimp.DBExec(db, "drop table if exists dtm_busi.user_account_log")
	logger.FatalIfError(err)
	_, err = dtmimp.DBExec(db, `create table if not exists dtm_busi.user_account_log (
	id      INT(11) AUTO_INCREMENT PRIMARY KEY,
	user_id INT(11) NOT NULL,
	delta DECIMAL(11, 2) not null,
	gid varchar(45) not null,
	branch_id varchar(45) not null,
	op varchar(45) not null,
	reason varchar(45),
	create_time datetime not null default now(),
	update_time datetime not null default now(),
	key(user_id),
	key(create_time)
)
`)
	logger.FatalIfError(err)
}

// StartSvr 1
func StartSvr() {
	app := dtmutil.GetGinApp()
	benchAddRoute(app)
	logger.Debugf("bench listening at %d", benchPort)
	go app.Run(fmt.Sprintf(":%s", benchPort))
}

func qsAdjustBalance(uid int, amount int, c *gin.Context) (interface{}, error) {
	if strings.Contains(mode, "empty") || sqls == 0 {
		return dtmcli.MapSuccess, nil
	}
	tb := dtmimp.TransBaseFromQuery(c.Request.URL.Query())
	f := func(tx *sql.Tx) error {
		for i := 0; i < sqls; i++ {
			_, err := dtmimp.DBExec(tx, "insert into dtm_busi.user_account_log(user_id, delta, gid, branch_id, op, reason)  values(?,?,?,?,?,?)",
				uid, amount, tb.Gid, c.Query("branch_id"), tb.TransType, fmt.Sprintf("inserted by dtm transaction %s %s", tb.Gid, c.Query("branch_id")))
			logger.FatalIfError(err)
			_, err = dtmimp.DBExec(tx, "update dtm_busi.user_account set balance = balance + ?, update_time = now() where user_id = ?", amount, uid)
			logger.FatalIfError(err)
		}
		return nil
	}
	if strings.Contains(mode, "barrier") {
		barrier, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		barrier.Call(txGet(), f)
	} else {
		tx := txGet()
		f(tx)
		err := tx.Commit()
		logger.FatalIfError(err)
	}

	return dtmcli.MapSuccess, nil
}

func benchAddRoute(app *gin.Engine) {
	app.POST(benchAPI+"/TransIn", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmimp.MustAtoi(c.Query("uid")), 1, c)
	}))
	app.POST(benchAPI+"/TransInCompensate", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmimp.MustAtoi(c.Query("uid")), -1, c)
	}))
	app.POST(benchAPI+"/TransOut", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmimp.MustAtoi(c.Query("uid")), -1, c)
	}))
	app.POST(benchAPI+"/TransOutCompensate", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmimp.MustAtoi(c.Query("uid")), 30, c)
	}))
	app.Any(benchAPI+"/reloadData", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		reloadData()
		mode = c.Query("m")
		s := c.Query("sqls")
		if s != "" {
			sqls = dtmimp.MustAtoi(s)
		}
		return nil, nil
	}))
	app.Any(benchAPI+"/bench", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		uid := (atomic.AddInt32(&uidCounter, 1)-1)%total + 1
		suid := fmt.Sprintf("%d", uid)
		suid2 := fmt.Sprintf("%d", total+1-uid)
		req := gin.H{}
		params := fmt.Sprintf("?uid=%s", suid)
		params2 := fmt.Sprintf("?uid=%s", suid2)
		logger.Debugf("mode: %s contains dtm: %t", mode, strings.Contains(mode, "dtm"))
		if strings.Contains(mode, "dtm") {
			saga := dtmcli.NewSaga(dtmutil.DefaultHttpServer, fmt.Sprintf("bench-%d", uid)).
				Add(benchBusi+"/TransOut"+params, benchBusi+"/TransOutCompensate"+params, req).
				Add(benchBusi+"/TransIn"+params2, benchBusi+"/TransInCompensate"+params2, req)
			saga.WaitResult = true
			err := saga.Submit()
			dtmimp.E2P(err)
		} else {
			_, err := dtmimp.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid2).Post(benchBusi + "/TransOut")
			dtmimp.E2P(err)
			_, err = dtmimp.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid).Post(benchBusi + "/TransIn")
			dtmimp.E2P(err)
		}
		return nil, nil
	}))
	app.Any(benchAPI+"/benchEmptyUrl", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		gid := shortuuid.New()
		req := gin.H{}
		saga := dtmcli.NewSaga(dtmutil.DefaultHttpServer, gid).
			Add("", "", req).
			Add("", "", req)
		saga.WaitResult = true
		err := saga.Submit()
		return nil, err
	}))
}

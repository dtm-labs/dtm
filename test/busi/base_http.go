/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// BusiAPI busi api prefix
	BusiAPI = "/api/busi"
	// BusiPort busi server port
	BusiPort = 8081
	// BusiGrpcPort busi server port
	BusiGrpcPort = 58081
)

type setupFunc func(*gin.Engine)

var setupFuncs = map[string]setupFunc{}

// Busi busi service url prefix
var Busi = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

// SleepCancelHandler 1
type SleepCancelHandler func(c *gin.Context) interface{}

var sleepCancelHandler SleepCancelHandler

// SetSleepCancelHandler 1
func SetSleepCancelHandler(handler SleepCancelHandler) {
	sleepCancelHandler = handler
}

// WebHookResult 1
var WebHookResult gin.H

// BaseAppStartup base app startup
func BaseAppStartup() *gin.Engine {
	logger.Infof("examples starting")
	app := dtmutil.GetGinApp()
	app.Use(func(c *gin.Context) {
		v := MainSwitch.NextResult.Fetch()
		if v != "" {
			c.JSON(200, gin.H{"dtm_result": v})
			c.Abort()
			return
		}
		c.Next()
	})

	BaseAddRoute(app)
	addJrpcRoute(app)
	for k, v := range setupFuncs {
		logger.Debugf("initing %s", k)
		v(app)
	}
	return app
}

// RunHTTP will run http server
func RunHTTP(app *gin.Engine) {
	logger.Debugf("Starting busi at: %d", BusiPort)
	err := app.Run(fmt.Sprintf(":%d", BusiPort))
	dtmimp.FatalIfError(err)
}

// BaseAddRoute add base route handler
func BaseAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/workflow/resume", dtmutil.WrapHandler(func(ctx *gin.Context) interface{} {
		data, err := io.ReadAll(ctx.Request.Body)
		logger.FatalIfError(err)
		return workflow.ExecuteByQS(ctx.Request.URL.Query(), data)
	}))
	app.POST(BusiAPI+"/TransIn", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiAPI+"/TransOut", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInConfirm", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInConfirmResult.Fetch(), "", "TransInConfirm")
	}))
	app.POST(BusiAPI+"/TransOutConfirm", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "TransOutConfirm")
	}))
	app.POST(BusiAPI+"/TransInRevert", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInRevertResult.Fetch(), "", "TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutRevert", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutRevertResult.Fetch(), "", "TransOutRevert")
	}))
	app.POST(BusiAPI+"/TransInOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiAPI+"/TransOutOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInConfirmOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransInConfirmResult.Fetch(), "", "TransInConfirm")
	}))
	app.POST(BusiAPI+"/TransOutConfirmOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "TransOutConfirm")
	}))
	app.POST(BusiAPI+"/TransInRevertOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransInRevertResult.Fetch(), "", "TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutRevertOld", oldWrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusinessCompatible(c, MainSwitch.TransOutRevertResult.Fetch(), "", "TransOutRevert")
	}))

	app.GET(BusiAPI+"/QueryPrepared", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		logger.Debugf("%s QueryPrepared", c.Query("gid"))
		return string2DtmError(dtmimp.OrString(MainSwitch.QueryPreparedResult.Fetch(), dtmcli.ResultSuccess))
	}))
	app.GET(BusiAPI+"/QueryPreparedB", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		logger.Debugf("%s QueryPreparedB", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		db := dbGet().ToSQLDB()
		return bb.QueryPrepared(db)
	}))
	app.GET(BusiAPI+"/RedisQueryPrepared", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		logger.Debugf("%s RedisQueryPrepared", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		return bb.RedisQueryPrepared(RedisGet(), 86400)
	}))
	app.GET(BusiAPI+"/MongoQueryPrepared", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		logger.Debugf("%s MongoQueryPrepared", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		return bb.MongoQueryPrepared(MongoGet())
	}))
	app.POST(BusiAPI+"/TransInXa", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return dtmcli.XaLocalTransaction(c.Request.URL.Query(), BusiConf, func(db *sql.DB, xa *dtmcli.Xa) error {
			return SagaAdjustBalance(db, TransInUID, reqFrom(c).Amount, reqFrom(c).TransInResult)
		})
	}))
	app.POST(BusiAPI+"/TransOutXa", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return dtmcli.XaLocalTransaction(c.Request.URL.Query(), BusiConf, func(db *sql.DB, xa *dtmcli.Xa) error {
			return SagaAdjustBalance(db, TransOutUID, -reqFrom(c).Amount, reqFrom(c).TransOutResult)
		})
	}))
	app.POST(BusiAPI+"/TransOutTimeout", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInTccNested", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		tcc, err := dtmcli.TccFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		logger.Debugf("TransInTccNested ")
		resp, err := tcc.CallBranch(&ReqHTTP{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		if err != nil {
			return err
		}
		return resp
	}))
	app.POST(BusiAPI+"/TransOutXaGorm", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return dtmcli.XaLocalTransaction(c.Request.URL.Query(), BusiConf, func(db *sql.DB, xa *dtmcli.Xa) error {
			if reqFrom(c).TransOutResult == dtmcli.ResultFailure {
				return dtmcli.ErrFailure
			}
			var dia gorm.Dialector
			if dtmcli.GetCurrentDBType() == dtmcli.DBTypeMysql {
				dia = mysql.New(mysql.Config{Conn: db})
			} else if dtmcli.GetCurrentDBType() == dtmcli.DBTypePostgres {
				dia = postgres.New(postgres.Config{Conn: db})
			}
			gdb, err := gorm.Open(dia, &gorm.Config{})
			if err != nil {
				return err
			}
			dbr := gdb.Exec("update dtm_busi.user_account set balance=balance-? where user_id=?", reqFrom(c).Amount, TransOutUID)
			return dbr.Error
		})
	}))

	app.POST(BusiAPI+"/TestPanic", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		if c.Query("panic_error") != "" {
			panic(errors.New("panic_error"))
		} else if c.Query("panic_string") != "" {
			panic("panic_string")
		}
		return nil
	}))
	app.POST(BusiAPI+"/SleepCancel", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		return sleepCancelHandler(c)
	}))
	app.POST(BusiAPI+"/TransOutHeaderYes", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		h := c.GetHeader("test_header")
		if h == "" {
			return errors.New("no test_header found in TransOutHeaderYes")
		}
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransOutHeaderNo", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		h := c.GetHeader("test_header")
		if h != "" {
			return errors.New("test_header found in TransOutHeaderNo")
		}
		return nil
	}))

	retryNums := 3
	app.POST(BusiAPI+"/TransInRetry", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		if retryNums != 0 {
			retryNums--
			return fmt.Errorf(("should be retried"))
		}
		retryNums = 3
		return nil
	}))
	app.POST(BusiAPI+"/AlertWebHook", dtmutil.WrapHandler(func(ctx *gin.Context) interface{} {
		err := ctx.BindJSON(&WebHookResult)
		dtmimp.FatalIfError(err)
		if strings.Contains(WebHookResult["gid"].(string), "Error") {
			return errors.New("gid contains 'Error', so return error")
		}
		return nil
	}))
}

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

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmutil"
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

// XaClient 1
var XaClient *dtmcli.XaClient

// SleepCancelHandler 1
type SleepCancelHandler func(c *gin.Context) interface{}

var sleepCancelHandler SleepCancelHandler

// SetSleepCancelHandler 1
func SetSleepCancelHandler(handler SleepCancelHandler) {
	sleepCancelHandler = handler
}

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
	var err error
	XaClient, err = dtmcli.NewXaClient(dtmutil.DefaultHTTPServer, BusiConf, Busi+"/xa", func(path string, xa *dtmcli.XaClient) {
		app.POST(path, dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
			return xa.HandleCallback(c.Query("gid"), c.Query("branch_id"), c.Query("op"))
		}))
	})
	logger.FatalIfError(err)

	BaseAddRoute(app)
	for k, v := range setupFuncs {
		logger.Debugf("initing %s", k)
		v(app)
	}
	logger.Debugf("Starting busi at: %d", BusiPort)
	go func() {
		_ = app.Run(fmt.Sprintf(":%d", BusiPort))
	}()
	return app
}

// BaseAddRoute add base route handler
func BaseAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/TransIn", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiAPI+"/TransOut", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInConfirm", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInConfirmResult.Fetch(), "", "TransInConfirm")
	}))
	app.POST(BusiAPI+"/TransOutConfirm", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "TransOutConfirm")
	}))
	app.POST(BusiAPI+"/TransInRevert", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return handleGeneralBusiness(c, MainSwitch.TransInRevertResult.Fetch(), "", "TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutRevert", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
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

	app.GET(BusiAPI+"/QueryPrepared", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		logger.Debugf("%s QueryPrepared", c.Query("gid"))
		return dtmcli.String2DtmError(dtmimp.OrString(MainSwitch.QueryPreparedResult.Fetch(), dtmcli.ResultSuccess))
	}))
	app.GET(BusiAPI+"/QueryPreparedB", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		logger.Debugf("%s QueryPreparedB", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		db := dbGet().ToSQLDB()
		return bb.QueryPrepared(db)
	}))
	app.GET(BusiAPI+"/RedisQueryPrepared", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		logger.Debugf("%s RedisQueryPrepared", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		return bb.RedisQueryPrepared(RedisGet(), 86400)
	}))
	app.GET(BusiAPI+"/MongoQueryPrepared", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		logger.Debugf("%s MongoQueryPrepared", c.Query("gid"))
		bb := MustBarrierFromGin(c)
		return bb.MongoQueryPrepared(MongoGet())
	}))
	app.POST(BusiAPI+"/TransInXa", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
			return SagaAdjustBalance(db, TransInUID, reqFrom(c).Amount, reqFrom(c).TransInResult)
		})
	}))
	app.POST(BusiAPI+"/TransOutXa", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
			return SagaAdjustBalance(db, TransOutUID, reqFrom(c).Amount, reqFrom(c).TransOutResult)
		})
	}))

	app.POST(BusiAPI+"/TransInTccNested", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		tcc, err := dtmcli.TccFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		logger.Debugf("TransInTccNested ")
		resp, err := tcc.CallBranch(&TransReq{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		if err != nil {
			return err
		}
		return resp
	}))
	app.POST(BusiAPI+"/TransOutXaGorm", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
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
	app.POST(BusiAPI+"/TccBSleepCancel", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return sleepCancelHandler(c)
	}))
	app.POST(BusiAPI+"/TransOutHeaderYes", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		h := c.GetHeader("test_header")
		if h == "" {
			return errors.New("no test_header found in TransOutHeaderYes")
		}
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransOutHeaderNo", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		h := c.GetHeader("test_header")
		if h != "" {
			return errors.New("test_header found in TransOutHeaderNo")
		}
		return nil
	}))
}

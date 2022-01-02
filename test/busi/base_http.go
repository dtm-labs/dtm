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
var Busi string = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

var XaClient *dtmcli.XaClient = nil

type SleepCancelHandler func(c *gin.Context) (interface{}, error)

var sleepCancelHandler SleepCancelHandler = nil

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
	XaClient, err = dtmcli.NewXaClient(dtmutil.DefaultHttpServer, BusiConf, Busi+"/xa", func(path string, xa *dtmcli.XaClient) {
		app.POST(path, dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
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
	go app.Run(fmt.Sprintf(":%d", BusiPort))

	return app
}

// BaseAddRoute add base route handler
func BaseAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/TransIn", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiAPI+"/TransOut", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInConfirm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInConfirmResult.Fetch(), "", "TransInConfirm")
	}))
	app.POST(BusiAPI+"/TransOutConfirm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "TransOutConfirm")
	}))
	app.POST(BusiAPI+"/TransInRevert", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInRevertResult.Fetch(), "", "TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutRevert", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutRevertResult.Fetch(), "", "TransOutRevert")
	}))
	app.GET(BusiAPI+"/CanSubmit", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		logger.Debugf("%s CanSubmit", c.Query("gid"))
		return dtmimp.OrString(MainSwitch.CanSubmitResult.Fetch(), dtmcli.ResultSuccess), nil
	}))
	app.POST(BusiAPI+"/TransInXa", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		err := XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
			return sagaAdjustBalance(db, transInUID, reqFrom(c).Amount, reqFrom(c).TransInResult)
		})
		return error2Resp(err)
	}))
	app.POST(BusiAPI+"/TransOutXa", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		err := XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
			return sagaAdjustBalance(db, transOutUID, reqFrom(c).Amount, reqFrom(c).TransOutResult)
		})
		return error2Resp(err)
	}))

	app.POST(BusiAPI+"/TransInTccParent", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		tcc, err := dtmcli.TccFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		logger.Debugf("TransInTccParent ")
		return tcc.CallBranch(&TransReq{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutXaGorm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		err := XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) error {
			if reqFrom(c).TransOutResult == dtmcli.ResultFailure {
				return dtmcli.ErrFailure
			}
			var dia gorm.Dialector = nil
			if dtmcli.GetCurrentDBType() == dtmcli.DBTypeMysql {
				dia = mysql.New(mysql.Config{Conn: db})
			} else if dtmcli.GetCurrentDBType() == dtmcli.DBTypePostgres {
				dia = postgres.New(postgres.Config{Conn: db})
			}
			gdb, err := gorm.Open(dia, &gorm.Config{})
			if err != nil {
				return err
			}
			dbr := gdb.Exec("update dtm_busi.user_account set balance=balance-? where user_id=?", reqFrom(c).Amount, transOutUID)
			return dbr.Error
		})
		return error2Resp(err)
	}))

	app.POST(BusiAPI+"/TestPanic", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		if c.Query("panic_error") != "" {
			panic(errors.New("panic_error"))
		} else if c.Query("panic_string") != "" {
			panic("panic_string")
		}
		return "SUCCESS", nil
	}))
	app.POST(BusiAPI+"/TccBSleepCancel", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return sleepCancelHandler(c)
	}))
	app.POST(BusiAPI+"/TransOutHeaderYes", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		h := c.GetHeader("test_header")
		if h == "" {
			return nil, errors.New("no test_header found in TransOutHeaderYes")
		}
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransOutHeaderNo", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
		h := c.GetHeader("test_header")
		if h != "" {
			return nil, errors.New("test_header found in TransOutHeaderNo")
		}
		return dtmcli.MapSuccess, nil
	}))
}

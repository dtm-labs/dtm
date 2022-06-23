/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
)

// GetGinApp init and return gin
func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())
	app.Use(func(c *gin.Context) {
		body := ""
		if c.Request.Body != nil {
			rb, err := c.GetRawData()
			dtmimp.E2P(err)
			if len(rb) > 0 {
				body = string(rb)
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			}
		}
		logger.Debugf("begin %s %s body: %s", c.Request.Method, c.Request.URL, body)
		c.Next()
	})
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, map[string]interface{}{"msg": "pong"}) })
	return app
}

// WrapHandler used by examples. much more simpler than WrapHandler2
func WrapHandler(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		ret := fn(c)
		status, res := dtmcli.Result2HttpJSON(ret)

		b, _ := json.Marshal(res)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		}
		c.JSON(status, res)
	}
}

// WrapHandler2 wrap a function to be the handler of gin request
// used by dtmsvr
func WrapHandler2(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		var err error
		r := func() interface{} {
			defer dtmimp.P2E(&err)
			return fn(c)
		}()

		status := http.StatusOK

		// in dtm test/busi, there are some functions, which will return a resty response
		// pass resty response as gin's response
		if resp, ok := r.(*resty.Response); ok {
			b := resp.Body()
			status = resp.StatusCode()
			r = nil
			err = json.Unmarshal(b, &r)
		}

		// error maybe returned in r, assign it to err
		if ne, ok := r.(error); ok && err == nil {
			err = ne
		}

		// if err != nil || r == nil. then set the status and dtm_result
		// dtm_result is for compatible with version lower than v1.10
		// when >= v1.10, result test should base on status, not dtm_result.
		result := map[string]interface{}{}
		if err != nil {
			if errors.Is(err, dtmcli.ErrFailure) {
				status = http.StatusConflict
				result["dtm_result"] = dtmcli.ResultFailure
			} else if errors.Is(err, dtmcli.ErrOngoing) {
				status = http.StatusTooEarly
				result["dtm_result"] = dtmcli.ResultOngoing
			} else if err != nil {
				status = http.StatusInternalServerError
			}
			result["message"] = err.Error()
			r = result
		} else if r == nil {
			result["dtm_result"] = dtmcli.ResultSuccess
			r = result
		}

		b, _ := json.Marshal(r)
		cont := string(b)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, cont)
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, cont)
		}
		c.JSON(status, r)
	}
}

// MustGetwd must version of os.Getwd
func MustGetwd() string {
	wd, err := os.Getwd()
	dtmimp.E2P(err)
	return wd
}

// GetSQLDir get sql scripts dir, used in test
func GetSQLDir() string {
	wd := MustGetwd()
	if filepath.Base(wd) == "test" {
		wd = filepath.Dir(wd)
	}
	return wd + "/sqls"
}

// RecoverPanic execs recovery operation
func RecoverPanic(err *error) {
	if x := recover(); x != nil {
		e := dtmimp.AsError(x)
		if err != nil {
			*err = e
		}
	}
}

// GetNextTime gets next time from second
func GetNextTime(second int64) *time.Time {
	next := time.Now().Add(time.Duration(second) * time.Second)
	return &next
}

// RunSQLScript 1
func RunSQLScript(conf dtmcli.DBConf, script string, skipDrop bool) {
	con, err := dtmimp.StandaloneDB(conf)
	logger.FatalIfError(err)
	defer func() { _ = con.Close() }()
	content, err := ioutil.ReadFile(script)
	logger.FatalIfError(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || (skipDrop && strings.Contains(s, "drop")) {
			continue
		}
		_, err = dtmimp.DBExec(conf.Driver, con, s)
		logger.FatalIfError(err)
		logger.Infof("sql scripts finished: %s", s)
	}
}

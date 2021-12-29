/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
		logger.Debugf("begin %s %s query: %s body: %s", c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)
		c.Next()
	})
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, map[string]interface{}{"msg": "pong"}) })
	return app
}

// WrapHandler name is clear
func WrapHandler(fn func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		r, err := func() (r interface{}, rerr error) {
			defer dtmimp.P2E(&rerr)
			return fn(c)
		}()
		var b = []byte{}
		if resp, ok := r.(*resty.Response); ok { // 如果是response，则取出body直接处理
			b = resp.Body()
		} else if err == nil {
			b, err = json.Marshal(r)
		}

		if err != nil {
			logger.Errorf("%2dms 500 %s %s %s %s", time.Since(began).Milliseconds(), err.Error(), c.Request.Method, c.Request.RequestURI, string(b))
			c.JSON(500, map[string]interface{}{"code": 500, "message": err.Error()})
		} else {
			logger.Infof("%2dms 200 %s %s %s", time.Since(began).Milliseconds(), c.Request.Method, c.Request.RequestURI, string(b))
			c.Status(200)
			c.Writer.Header().Add("Content-Type", "application/json")
			_, err = c.Writer.Write(b)
			dtmimp.E2P(err)
		}
	}
}

// MustGetwd must version of os.Getwd
func MustGetwd() string {
	wd, err := os.Getwd()
	dtmimp.E2P(err)
	return wd
}

// GetSqlDir 获取调用该函数的caller源代码的目录，主要用于测试时，查找相关文件
func GetSqlDir() string {
	wd := MustGetwd()
	if filepath.Base(wd) == "test" {
		wd = filepath.Dir(wd)
	}
	return wd + "/sqls"
}

func RecoverPanic(err *error) {
	if x := recover(); x != nil {
		e := dtmimp.AsError(x)
		if err != nil {
			*err = e
		}
	}
}

func GetNextTime(second int64) *time.Time {
	next := time.Now().Add(time.Duration(second) * time.Second)
	return &next
}

// RunSQLScript 1
func RunSQLScript(conf dtmcli.DBConf, script string, skipDrop bool) {
	con, err := dtmimp.StandaloneDB(conf)
	logger.FatalIfError(err)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	logger.FatalIfError(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || (skipDrop && strings.Contains(s, "drop")) {
			continue
		}
		_, err = dtmimp.DBExec(con, s)
		logger.FatalIfError(err)
		logger.Infof("sql scripts finished: %s", s)
	}
}

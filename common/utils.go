/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

// GetGinApp init and return gin
func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
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
		began := time.Now()
		dtmimp.Logf("begin %s %s query: %s body: %s", c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)
		c.Next()
		dtmimp.Logf("used %d ms %s %s query: %s body: %s", time.Since(began).Milliseconds(), c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)

	})
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, map[string]interface{}{"msg": "pong"}) })
	return app
}

// WrapHandler name is clear
func WrapHandler(fn func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			dtmimp.Logf("status: 500, code: 500 message: %s", err.Error())
			c.JSON(500, map[string]interface{}{"code": 500, "message": err.Error()})
		} else {
			dtmimp.Logf("status: 200, content: %s", string(b))
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

// GetCallerCodeDir 获取调用该函数的caller源代码的目录，主要用于测试时，查找相关文件
func GetCallerCodeDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

func RecoverPanic(err *error) {
	if x := recover(); x != nil {
		e := dtmimp.AsError(x)
		if err != nil {
			err = &e
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
	dtmimp.FatalIfError(err)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	dtmimp.FatalIfError(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || (skipDrop && strings.Contains(s, "drop")) {
			continue
		}
		_, err = dtmimp.DBExec(con, s)
		dtmimp.FatalIfError(err)
	}
}

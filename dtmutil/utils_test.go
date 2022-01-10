/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmutil

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGin(t *testing.T) {
	app := GetGinApp()
	app.GET("/api/sample", WrapHandler2(func(c *gin.Context) interface{} {
		return 1
	}))
	app.GET("/api/error", WrapHandler2(func(c *gin.Context) interface{} {
		return errors.New("err1")
	}))
	getResultString := func(api string, body io.Reader) string {
		req, _ := http.NewRequest("GET", api, body)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		return w.Body.String()
	}
	assert.Equal(t, "{\"msg\":\"pong\"}", getResultString("/api/ping", nil))
	assert.Equal(t, "1", getResultString("/api/sample", nil))
	assert.Equal(t, "{\"message\":\"err1\"}", getResultString("/api/error", strings.NewReader("{}")))
}

func TestFuncs(t *testing.T) {
	wd := MustGetwd()
	assert.NotEqual(t, "", wd)

	dir1 := GetSQLDir()
	assert.Equal(t, true, strings.HasSuffix(dir1, "/sqls"))

}

func TestRecoverPanic(t *testing.T) {
	err := func() (rerr error) {
		defer RecoverPanic(&rerr)
		panic(fmt.Errorf("an error"))
	}()
	assert.Equal(t, "an error", err.Error())
}

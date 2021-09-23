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
)

// GetGinApp init and return gin
func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	app.Use(func(c *gin.Context) {
		body := ""
		if c.Request.Body != nil {
			rb, err := c.GetRawData()
			dtmcli.E2P(err)
			if len(rb) > 0 {
				body = string(rb)
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			}
		}
		began := time.Now()
		dtmcli.Logf("begin %s %s query: %s body: %s", c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)
		c.Next()
		dtmcli.Logf("used %d ms %s %s query: %s body: %s", time.Since(began).Milliseconds(), c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)

	})
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, dtmcli.M{"msg": "pong"}) })
	return app
}

// WrapHandler name is clear
func WrapHandler(fn func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := fn(c)
		var b = []byte{}
		if resp, ok := r.(*resty.Response); ok { // 如果是response，则取出body直接处理
			b = resp.Body()
		} else if err == nil {
			b, err = json.Marshal(r)
		}
		if err != nil {
			dtmcli.Logf("status: 500, code: 500 message: %s", err.Error())
			c.JSON(500, dtmcli.M{"code": 500, "message": err.Error()})
		} else {
			dtmcli.Logf("status: 200, content: %s", string(b))
			c.Status(200)
			c.Writer.Header().Add("Content-Type", "application/json")
			_, err = c.Writer.Write(b)
			dtmcli.E2P(err)
		}
	}
}

// MustGetwd must version of os.Getwd
func MustGetwd() string {
	wd, err := os.Getwd()
	dtmcli.E2P(err)
	return wd
}

// GetCallerCodeDir 获取调用该函数的caller源代码的目录，主要用于测试时，查找相关文件
func GetCallerCodeDir() string {
	_, file, _, _ := runtime.Caller(1)
	wd := MustGetwd()
	if strings.HasSuffix(wd, "/test") {
		wd = filepath.Dir(wd)
	}
	return wd + "/" + filepath.Base(filepath.Dir(file))
}

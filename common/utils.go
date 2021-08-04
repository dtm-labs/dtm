package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	yaml "gopkg.in/yaml.v2"
)

// GetGinApp init and return gin
func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	app.Use(func(c *gin.Context) {
		body := ""
		if c.Request.Body != nil {
			rb, err := c.GetRawData()
			E2P(err)
			if len(rb) > 0 {
				body = string(rb)
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			}
		}
		began := time.Now()
		Logf("begin %s %s query: %s body: %s", c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)
		c.Next()
		Logf("used %d ms %s %s query: %s body: %s", time.Since(began).Milliseconds(), c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)

	})
	app.Any("/api/ping", func(c *gin.Context) { c.JSON(200, M{"msg": "pong"}) })
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
			Logf("status: 500, code: 500 message: %s", err.Error())
			c.JSON(500, M{"code": 500, "message": err.Error()})
		} else {
			Logf("status: 200, content: %s", string(b))
			c.Status(200)
			c.Writer.Header().Add("Content-Type", "application/json")
			_, err = c.Writer.Write(b)
			E2P(err)
		}
	}
}

// InitConfig init config
func InitConfig(dir string, config interface{}) {
	cont, err := ioutil.ReadFile(dir + "/conf.yml")
	if err != nil {
		cont, err = ioutil.ReadFile(dir + "/conf.sample.yml")
	}
	Logf("cont is: \n%s", string(cont))
	E2P(err)
	err = yaml.Unmarshal(cont, config)
	E2P(err)
}

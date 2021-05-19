package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func OrString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

var gNode *snowflake.Node = nil

func GenGid() string {
	return gNode.Generate().Base58()
}

func init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	gNode = node
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func If(condition bool, trueObj interface{}, falseObj interface{}) interface{} {
	if condition {
		return trueObj
	}
	return falseObj
}

func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	PanicIfError(err)
	return b
}

func MustMarshalString(v interface{}) string {
	return string(MustMarshal(v))
}

func MustUnmarshalString(s string, obj interface{}) {
	err := json.Unmarshal([]byte(s), obj)
	PanicIfError(err)
}

func MustRemarshal(from interface{}, to interface{}) {
	b, err := json.Marshal(from)
	PanicIfError(err)
	err = json.Unmarshal(b, to)
	PanicIfError(err)
}

var RestyClient = resty.New()

func init() {
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logrus.Printf("requesting: %s %s %v", r.Method, r.URL, r.Body)
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logrus.Printf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}

func GetGinApp() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	app.Use(func(c *gin.Context) {
		body := ""
		if c.Request.Method == "POST" {
			rb, err := c.GetRawData()
			if err != nil {
				logrus.Printf("GetRawData error: %s", err.Error())
			} else {
				body = string(rb)
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rb))
			}
		}
		began := time.Now()
		logrus.Printf("begin %s %s query: %s body: %s", c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)
		c.Next()
		logrus.Printf("used %d ms %s %s query: %s body: %s", time.Since(began).Milliseconds(), c.Request.Method, c.FullPath(), c.Request.URL.RawQuery, body)

	})
	return app
}

package dtmsvr

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Main() {
	logrus.Printf("start dtmsvr")
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
		logrus.Printf("used %d ms %s %s query: %s", time.Since(began).Milliseconds(), c.Request.Method, c.FullPath(), c.Request.URL.RawQuery)

	})
	AddRoute(app)
	// StartConsumePreparedMsg(1)
	StartConsumeCommitedMsg(1)
	logrus.Printf("dtmsvr listen at: 8080")
	go app.Run()
}

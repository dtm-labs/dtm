/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/entry"
	_ "github.com/dtm-labs/dtm/dtmsvr/microservices"
	"github.com/gin-gonic/gin"
)

// Version defines version info. It is set by -ldflags.
var Version string

func main() {
	app := entry.Main(&Version)
	if app != nil {
		addDashboard(app)
		select {}
	}
}
func addDashboard(app *gin.Engine) {
	app.GET("/dashboard/*name", proxyDashboard)
	app.GET("/@vite/*name", proxyDashboard)
	app.GET("/node_modules/*name", proxyDashboard)
	app.GET("/src/*name", proxyDashboard)
	app.GET("/@id/*name", proxyDashboard)
}

func proxyDashboard(c *gin.Context) {

	target := "127.0.0.1:5000"
	u := &url.URL{}
	u.Scheme = "http"
	u.Host = target
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("http: proxy error: %v", err)
		ret := fmt.Sprintf("http proxy error %v", err)
		_, _ = rw.Write([]byte(ret))
	}
	logger.Debugf("proxy dashboard to %s", target)
	proxy.ServeHTTP(c.Writer, c.Request)

}

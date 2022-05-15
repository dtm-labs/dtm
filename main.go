/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/entry"
	_ "github.com/dtm-labs/dtm/dtmsvr/microservices"
	"github.com/gin-gonic/gin"
)

// Version defines version info. It is set by -ldflags.
var Version string

func main() {
	app, conf := entry.Main(&Version)
	if app != nil {
		addDashboard(app, conf)
		select {}
	}
}

//go:embed dashboard/dist
var dashboard embed.FS

//go:embed dashboard/dist/index.html
var indexFile string

var target = "127.0.0.1:5000"

func getSub(f1 fs.FS, sub string) fs.FS {
	f2, err := fs.Sub(f1, sub)
	logger.FatalIfError(err)
	return f2
}
func addDashboard(app *gin.Engine, conf *config.ConfigType) {
	dist := getSub(dashboard, "dashboard/dist")
	_, err := dist.Open("index.html")
	if err == nil {
		app.StaticFS("/assets", http.FS(getSub(dist, "assets")))
		app.GET("/dashboard/*name", func(c *gin.Context) {
			c.Header("content-type", "text/html;charset=utf-8")
			c.String(200, indexFile)
		})
		logger.Infof("dashboard is served from dir 'dashboard/dist/'")
	} else {
		app.GET("/", proxyDashboard)
		app.GET("/dashboard/*name", proxyDashboard)
		app.GET("/@vite/*name", proxyDashboard)
		app.GET("/node_modules/*name", proxyDashboard)
		app.GET("/src/*name", proxyDashboard)
		app.GET("/@id/*name", proxyDashboard)
		logger.Infof("dashboard is proxied to %s", target)
	}
	logger.Infof("Dashboard is running at: http://localhost:%d", conf.HTTPPort)
}

func proxyDashboard(c *gin.Context) {

	u := &url.URL{}
	u.Scheme = "http"
	u.Host = target
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.Warnf("http: proxy error: %v", err)
		ret := fmt.Sprintf("http proxy error %v", err)
		_, _ = rw.Write([]byte(ret))
	}
	logger.Debugf("proxy dashboard to %s", target)
	proxy.ServeHTTP(c.Writer, c.Request)

}

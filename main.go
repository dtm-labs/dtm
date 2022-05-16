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
	"io/ioutil"
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
		addAdmin(app, conf)
		select {}
	}
}

//go:embed admin/dist
var admin embed.FS

var target = "admin.dtm.pub"

func getSub(f1 fs.FS, sub string) fs.FS {
	f2, err := fs.Sub(f1, sub)
	logger.FatalIfError(err)
	return f2
}
func addAdmin(app *gin.Engine, conf *config.ConfigType) {
	// for released dtm, serve admin from local files because the build output has been embed
	// for testing users, proxy admin to target because the build output has not been embed
	dist := getSub(admin, "admin/dist")
	index, err := dist.Open("index.html")
	if err == nil {
		defer index.Close()
		cont, err := ioutil.ReadAll(index)
		logger.FatalIfError(err)
		sfile := string(cont)
		renderIndex := func(c *gin.Context) {
			c.Header("content-type", "text/html;charset=utf-8")
			c.String(200, sfile)
		}
		app.StaticFS("/assets", http.FS(getSub(dist, "assets")))
		app.GET("/admin/*name", renderIndex)
		app.GET("/", renderIndex)
		logger.Infof("admin is served from dir 'admin/dist/'")
	} else {
		app.GET("/", proxyAdmin)
		app.GET("/admin/*name", proxyAdmin)
		logger.Infof("admin is proxied to %s", target)
	}
	logger.Infof("admin is running at: http://localhost:%d", conf.HTTPPort)
}

func proxyAdmin(c *gin.Context) {

	u := &url.URL{}
	u.Scheme = "http"
	u.Host = target
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.Warnf("http: proxy error: %v", err)
		ret := fmt.Sprintf("http proxy error %v", err)
		_, _ = rw.Write([]byte(ret))
	}
	logger.Debugf("proxy admin to %s", target)
	proxy.ServeHTTP(c.Writer, c.Request)

}

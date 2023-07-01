/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

// package main is the entry of dtm server
package main

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/entry"
	_ "github.com/dtm-labs/dtm/dtmsvr/microservices"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
)

// Version defines version info. It is set by -ldflags.
var Version string

func main() {
	app, conf := entry.Main(&Version)
	if app != nil {
		addAdmin(app, conf)
		q := make(chan os.Signal, 1)
		signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
		<-q
		logger.Infof("Shutdown dtm server...")
	}
}

//go:embed admin/dist
var admin embed.FS

var target = ""

func getSub(f1 fs.FS, subs ...string) fs.FS {
	var err error
	for _, sub := range subs {
		f1, err = fs.Sub(f1, sub)
		logger.FatalIfError(err)
	}
	return f1
}
func addAdmin(app *gin.Engine, conf *config.Type) {
	// for released dtm, serve admin from local files because the build output has been embed
	// for testing users, proxy admin to target because the build output has not been embed
	dist := getSub(admin, "admin", "dist")
	index, err := dist.Open("index.html")

	var router gin.IRoutes = app
	if len(conf.AdminBasePath) > 0 {
		router = app.Group(conf.AdminBasePath)
	}
	if err == nil {
		cont, err := io.ReadAll(index)
		logger.FatalIfError(err)
		_ = index.Close()
		cont = bytesTryReplaceIndex(cont, conf)

		renderIndex := func(c *gin.Context) {
			c.Data(200, "text/html; charset=utf-8", cont)
		}
		router.StaticFS("/assets", http.FS(getSub(dist, "assets")))
		router.GET("/admin/*name", renderIndex)
		router.GET("/", renderIndex)
		router.GET("/favicon.ico", func(ctx *gin.Context) {
			http.StripPrefix(conf.AdminBasePath, http.FileServer(http.FS(dist))).ServeHTTP(ctx.Writer, ctx.Request)
		})
		logger.Infof("admin is served from dir 'admin/dist/'")
	} else {
		router.GET("/", proxyAdmin(conf))
		router.GET("/assets/*name", proxyAdmin(conf))
		router.GET("/admin/*name", proxyAdmin(conf))
		lang := os.Getenv("LANG")
		if strings.HasPrefix(lang, "zh_CN") {
			target = "cn-admin.dtm.pub"
		} else {
			target = "admin.dtm.pub"
		}
		logger.Infof("admin is proxied to %s", target)
	}
	logger.Infof("admin is running at: http://localhost:%d%s", conf.HTTPPort, conf.AdminBasePath)
}

func proxyAdmin(conf *config.Type) func(c *gin.Context) {
	return func(c *gin.Context) {
		u := &url.URL{}
		u.Scheme = "http"
		u.Host = target
		proxy := httputil.NewSingleHostReverseProxy(u)
		originalDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			originalDirector(r)
			p := strings.TrimPrefix(r.URL.Path, conf.AdminBasePath)
			rp := strings.TrimPrefix(r.URL.RawPath, conf.AdminBasePath)
			r.URL.Path = p
			r.URL.RawPath = rp
		}
		proxy.Transport = &transport{RoundTripper: http.DefaultTransport, conf: conf}
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			logger.Warnf("http: proxy error: %v", err)
			ret := fmt.Sprintf("http proxy error %v", err)
			_, _ = rw.Write([]byte(ret))
		}
		logger.Debugf("proxy admin to %s", target)
		c.Request.Host = target
		proxy.ServeHTTP(c.Writer, c.Request)
	}

}

// bytesTryReplaceIndex replace index.html base path
func bytesTryReplaceIndex(source []byte, conf *config.Type) []byte {
	source = bytes.Replace(source, []byte("\"assets/"), []byte("\""+conf.AdminBasePath+"/assets/"), -1)
	source = bytes.Replace(source, []byte("PUBLIC-PATH-VARIABLE"), []byte(conf.AdminBasePath), -1)
	return source
}

type transport struct {
	http.RoundTripper
	conf *config.Type
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	//modify html only
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return resp, err
	}
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer func() {
			if tmpErr := reader.Close(); err == nil && tmpErr != nil {
				err = tmpErr
			}
		}()
	default:
		reader = resp.Body
	}
	delete(resp.Header, "Content-Encoding")
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	b = bytesTryReplaceIndex(b, t.conf)
	body := io.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return resp, nil
}

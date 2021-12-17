/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func addRoute(engine *gin.Engine) {
	engine.GET("/api/dtmsvr/newGid", common.WrapHandler(newGid))
	engine.POST("/api/dtmsvr/prepare", common.WrapHandler(prepare))
	engine.POST("/api/dtmsvr/submit", common.WrapHandler(submit))
	engine.POST("/api/dtmsvr/abort", common.WrapHandler(abort))
	engine.POST("/api/dtmsvr/registerBranch", common.WrapHandler(registerBranch))
	engine.POST("/api/dtmsvr/registerXaBranch", common.WrapHandler(registerBranch))  // compatible for old sdk
	engine.POST("/api/dtmsvr/registerTccBranch", common.WrapHandler(registerBranch)) // compatible for old sdk
	engine.GET("/api/dtmsvr/query", common.WrapHandler(query))
	engine.GET("/api/dtmsvr/all", common.WrapHandler(all))

	// add prometheus exporter
	h := promhttp.Handler()
	engine.GET("/api/metrics", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})
}

func newGid(c *gin.Context) (interface{}, error) {
	return map[string]interface{}{"gid": GenGid(), "dtm_result": dtmcli.ResultSuccess}, nil
}

func prepare(c *gin.Context) (interface{}, error) {
	return svcPrepare(TransFromContext(c))
}

func submit(c *gin.Context) (interface{}, error) {
	return svcSubmit(TransFromContext(c))
}

func abort(c *gin.Context) (interface{}, error) {
	return svcAbort(TransFromContext(c))
}

func registerBranch(c *gin.Context) (interface{}, error) {
	data := map[string]string{}
	err := c.BindJSON(&data)
	e2p(err)
	branch := TransBranch{
		Gid:      data["gid"],
		BranchID: data["branch_id"],
		Status:   dtmcli.StatusPrepared,
		BinData:  []byte(data["data"]),
	}
	return svcRegisterBranch(data["trans_type"], &branch, data)
}

func query(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	if gid == "" {
		return nil, errors.New("no gid specified")
	}
	trans := GetStore().FindTransGlobalStore(gid)
	branches := GetStore().FindBranches(gid)
	return map[string]interface{}{"transaction": trans, "branches": branches}, nil
}

func all(c *gin.Context) (interface{}, error) {
	position := c.Query("position")
	slimit := dtmimp.OrString(c.Query("limit"), "100")
	globals := GetStore().ScanTransGlobalStores(&position, int64(dtmimp.MustAtoi(slimit)))
	return map[string]interface{}{"transactions": globals, "next_position": position}, nil
}

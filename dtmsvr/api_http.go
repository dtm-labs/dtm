/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"errors"
	"strconv"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func addRoute(engine *gin.Engine) {
	engine.GET("/api/dtmsvr/version", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		return gin.H{"version": Version}
	}))
	engine.GET("/api/dtmsvr/newGid", dtmutil.WrapHandler2(newGid))
	engine.POST("/api/dtmsvr/prepare", dtmutil.WrapHandler2(prepare))
	engine.POST("/api/dtmsvr/submit", dtmutil.WrapHandler2(submit))
	engine.POST("/api/dtmsvr/abort", dtmutil.WrapHandler2(abort))
	engine.POST("/api/dtmsvr/forceStop", dtmutil.WrapHandler2(forceStop)) // change global status to failed can stop trigger (Use with caution in production environment)
	engine.POST("/api/dtmsvr/registerBranch", dtmutil.WrapHandler2(registerBranch))
	engine.POST("/api/dtmsvr/registerXaBranch", dtmutil.WrapHandler2(registerBranch))  // compatible for old sdk
	engine.POST("/api/dtmsvr/registerTccBranch", dtmutil.WrapHandler2(registerBranch)) // compatible for old sdk
	engine.POST("/api/dtmsvr/prepareWorkflow", dtmutil.WrapHandler2(prepareWorkflow))
	engine.GET("/api/dtmsvr/query", dtmutil.WrapHandler2(query))
	engine.GET("/api/dtmsvr/all", dtmutil.WrapHandler2(all))
	engine.GET("/api/dtmsvr/resetCronTime", dtmutil.WrapHandler2(resetCronTime))
	engine.GET("/api/dtmsvr/subscribe", dtmutil.WrapHandler2(subscribe))
	engine.GET("/api/dtmsvr/unsubscribe", dtmutil.WrapHandler2(unsubscribe))
	engine.DELETE("/api/dtmsvr/topic/:topicName", dtmutil.WrapHandler2(deleteTopic))
	engine.GET("/api/dtmsvr/scanKV", dtmutil.WrapHandler2(scanKV))
	engine.GET("/api/dtmsvr/queryKV", dtmutil.WrapHandler2(queryKV))
	engine.POST("/api/dtmsvr/resetNextCronTime", dtmutil.WrapHandler2(resetNextCronTime)) // one global trans only

	// add prometheus exporter
	h := promhttp.Handler()
	engine.GET("/api/metrics", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})
}

// NOTE: unique in storage, can customize the generation rules instead of using server-side generation, it will help with the tracking
func newGid(c *gin.Context) interface{} {
	return map[string]interface{}{"gid": GenGid(), "dtm_result": dtmcli.ResultSuccess}
}

func prepare(c *gin.Context) interface{} {
	return svcPrepare(TransFromContext(c))
}

func submit(c *gin.Context) interface{} {
	return svcSubmit(TransFromContext(c))
}

func abort(c *gin.Context) interface{} {
	return svcAbort(TransFromContext(c))
}

func forceStop(c *gin.Context) interface{} {
	return svcForceStop(TransFromContext(c))
}

func resetNextCronTime(c *gin.Context) interface{} {
	return svcResetNextCronTime(TransFromContext(c))
}

func registerBranch(c *gin.Context) interface{} {
	data := map[string]string{}
	err := c.BindJSON(&data)
	e2p(err)
	branch := TransBranch{
		Gid:      dtmimp.Escape(data["gid"]),
		BranchID: data["branch_id"],
		Status:   dtmcli.StatusPrepared,
		BinData:  []byte(data["data"]),
	}
	return svcRegisterBranch(data["trans_type"], &branch, data)
}

func query(c *gin.Context) interface{} {
	gid := c.Query("gid")
	if gid == "" {
		return errors.New("no gid specified")
	}
	trans := GetStore().FindTransGlobalStore(gid)
	branches := GetStore().FindBranches(gid)
	return map[string]interface{}{"transaction": trans, "branches": branches}
}

func prepareWorkflow(c *gin.Context) interface{} {
	trans, branches, err := svcPrepareWorkflow(TransFromContext(c))
	if err != nil {
		return err
	}
	return map[string]interface{}{"transaction": trans, "progresses": branches}
}

func all(c *gin.Context) interface{} {
	gid := c.Query("gid")
	position := c.Query("position")
	sLimit := dtmimp.OrString(c.Query("limit"), "100")

	var globals interface{}
	if len(gid) > 0 {
		find := GetStore().FindTransGlobalStore(gid)
		if find != nil {
			globals = []interface{}{*find}
		}
	} else {
		condition := storage.TransGlobalScanCondition{
			Status:          c.Query("status"),
			TransType:       c.Query("transType"),
			CreateTimeStart: stringTotime(c.Query("createTimeStart")),
			CreateTimeEnd:   stringTotime(c.Query("createTimeEnd")),
		}
		globals = GetStore().ScanTransGlobalStores(&position, int64(dtmimp.MustAtoi(sLimit)), condition)
	}
	return map[string]interface{}{"transactions": globals, "next_position": position}
}

func stringTotime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	return time.Unix(int64(dtmimp.MustAtoi(timeStr))/1000, 0)
}

// unfinished transactions need to be retried as soon as possible after business downtime is recovered
func resetCronTime(c *gin.Context) interface{} {
	sTimeoutSecond := dtmimp.OrString(c.Query("timeout"), strconv.FormatInt(3*conf.TimeoutToFail, 10))
	sLimit := dtmimp.OrString(c.Query("limit"), "100")
	timeout := time.Duration(dtmimp.MustAtoi(sTimeoutSecond)) * time.Second

	succeedCount, hasRemaining, err := GetStore().ResetCronTime(timeout, int64(dtmimp.MustAtoi(sLimit)))
	if err != nil {
		return err
	}
	return map[string]interface{}{"has_remaining": hasRemaining, "succeed_count": succeedCount}
}

func scanKV(c *gin.Context) interface{} {
	cat := c.DefaultQuery("cat", "")
	position := c.Query("position")
	sLimit := dtmimp.OrString(c.Query("limit"), "100")
	kv := GetStore().ScanKV(cat, &position, int64(dtmimp.MustAtoi(sLimit)))
	return map[string]interface{}{"kv": kv, "next_position": position}
}

func queryKV(c *gin.Context) interface{} {
	cat := c.DefaultQuery("cat", "")
	key := c.DefaultQuery("key", "")

	kv := GetStore().FindKV(cat, key)
	return map[string]interface{}{"kv": kv}
}

func subscribe(c *gin.Context) interface{} {
	topic := c.Query("topic")
	url := c.Query("url")
	remark := c.Query("remark")

	return Subscribe(topic, url, remark)
}

func unsubscribe(c *gin.Context) interface{} {
	topic := c.Query("topic")
	url := c.Query("url")

	return Unsubscribe(topic, url)
}

func deleteTopic(c *gin.Context) interface{} {
	topic := c.Param("topicName")
	if topic == "" {
		return errors.New("empty topic")
	}

	return GetStore().DeleteKV(topicsCat, topic)
}

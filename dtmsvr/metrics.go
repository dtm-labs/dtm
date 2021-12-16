/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
)

var (
	processTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dtm_server_process_total",
		Help: "All request received by dtm",
	},
		[]string{"type", "api", "status"})

	responseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dtm_server_response_duration",
		Help: "The request durations of a dtm server api",
	},
		[]string{"type", "api"})

	transactionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dtm_transaction_process_total",
		Help: "All transactions processed by dtm",
	},
		[]string{"model", "gid", "status"})

	branchTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dtm_branch_process_total",
		Help: "All branches processed by dtm",
	},
		[]string{"model", "gid", "branchid", "branchtype", "status"})
)

func httpMetrics(app *gin.Engine) *gin.Engine {
	app.Use(func(c *gin.Context) {
		api := extractFromPath(c.Request.RequestURI)
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			responseTime.WithLabelValues("http", api).Observe(v)
		}))
		defer timer.ObserveDuration()
		c.Next()
		status := c.Writer.Status()
		if status >= 400 {
			processTotal.WithLabelValues("http", api, "fail").Inc()
		} else {
			processTotal.WithLabelValues("http", api, "ok").Inc()
		}
	})
	return app
}

func grpcMetrics(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	api := extractFromPath(info.FullMethod)
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		responseTime.WithLabelValues("grpc", api).Observe(v)
	}))
	defer timer.ObserveDuration()
	m, err := handler(ctx, req)
	if err != nil {
		processTotal.WithLabelValues("grpc", api, "fail").Inc()
	} else {
		processTotal.WithLabelValues("grpc", api, "ok").Inc()
	}
	return m, err
}

func transactionMetrics(global *TransGlobal, status bool) {
	if status {
		transactionTotal.WithLabelValues(global.TransType, global.Gid, "ok").Inc()
	} else {
		transactionTotal.WithLabelValues(global.TransType, global.Gid, "fail").Inc()
	}
}

func branchMetrics(global *TransGlobal, branch *TransBranch, status bool) {
	if status {
		branchTotal.WithLabelValues(global.TransType, global.Gid, branch.BranchID, branch.Op, "ok").Inc()
	} else {
		branchTotal.WithLabelValues(global.TransType, global.Gid, branch.BranchID, branch.Op, "fail").Inc()
	}
}

func extractFromPath(val string) string {
	strs := strings.Split(val, "/")
	return strings.ToLower(strs[len(strs)-1])
}

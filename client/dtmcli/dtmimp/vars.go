/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"errors"
	"sync"
	"time"

	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	"github.com/go-resty/resty/v2"
)

// ErrFailure error of FAILURE
var ErrFailure = errors.New("FAILURE")

// ErrOngoing error of ONGOING
var ErrOngoing = errors.New("ONGOING")

// ErrDuplicated error of DUPLICATED for only msg
// if QueryPrepared executed before call. then DoAndSubmit return this error
var ErrDuplicated = errors.New("DUPLICATED")

// MapSuccess HTTP result of SUCCESS
var MapSuccess = map[string]interface{}{"dtm_result": ResultSuccess}

// MapFailure HTTP result of FAILURE
var MapFailure = map[string]interface{}{"dtm_result": ResultFailure}

// BarrierTableName the table name of barrier table
var BarrierTableName = "dtm_barrier.barrier"

var restyClients sync.Map

// GetRestyClient2 will return a resty client with timeout set
func GetRestyClient2(timeout time.Duration) *resty.Client {
	cli, ok := restyClients.Load(timeout)
	if !ok {
		client := resty.New()
		if timeout != 0 {
			client.SetTimeout(timeout)
		}
		AddRestyMiddlewares(client)
		restyClients.Store(timeout, client)
		cli = client
	}
	return cli.(*resty.Client)
}

// AddRestyMiddlewares will add the middlewares used by dtm
func AddRestyMiddlewares(client *resty.Client) {
	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		old := r.URL
		r.URL = MayReplaceLocalhost(r.URL)
		ms := dtmdriver.Middlewares.HTTP
		var err error
		for i := 0; i < len(ms) && err == nil; i++ {
			err = ms[i](c, r)
		}
		logger.Debugf("requesting: %s %s %s resolved: %s err: %v", r.Method, old, MustMarshalString(r.Body), r.URL, err)
		return err
	})
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logger.Debugf("requested: %d %s %s %s", resp.StatusCode(), r.Method, r.URL, resp.String())
		return nil
	})
}

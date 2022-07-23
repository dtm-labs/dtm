/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"errors"

	"github.com/dtm-labs/dtm/client/dtmcli/logger"
	"github.com/dtm-labs/dtmdriver"
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

// RestyClient the resty object
var RestyClient = resty.New()

// PassthroughHeaders will be passed to every sub-trans call
var PassthroughHeaders = []string{}

// BarrierTableName the table name of barrier table
var BarrierTableName = "dtm_barrier.barrier"

// AddRestyMiddlewares will add the middlewares used by dtm
func AddRestyMiddlewares(client *resty.Client) {
	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logger.Debugf("requesting: %s %s %s resolved: %s", r.Method, r.URL, MustMarshalString(r.Body), r.URL)
		r.URL = MayReplaceLocalhost(r.URL)
		ms := dtmdriver.Middlewares.HTTP
		var err error
		for i := 0; i < len(ms) && err == nil; i++ {
			err = ms[i](c, r)
		}
		return err
	})
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logger.Debugf("requested: %d %s %s %s", resp.StatusCode(), r.Method, r.URL, resp.String())
		return nil
	})
}

func init() {
	AddRestyMiddlewares(RestyClient)
}

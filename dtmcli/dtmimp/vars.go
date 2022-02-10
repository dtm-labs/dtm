/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"errors"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/go-resty/resty/v2"
)

// ErrFailure error of FAILURE
var ErrFailure = errors.New("FAILURE")

// ErrOngoing error of ONGOING
var ErrOngoing = errors.New("ONGOING")

// ErrDuplicated error of DUPLICATED for only msg
// if QueryPrepared executed before call. then DoAndSubmit return this error
var ErrDuplicated = errors.New("DUPLICATED")

// XaSQLTimeoutMs milliseconds for Xa sql to timeout
var XaSQLTimeoutMs = 15000

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

func init() {
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.URL = MayReplaceLocalhost(r.URL)
		logger.Debugf("requesting: %s %s %s", r.Method, r.URL, MustMarshalString(r.Body))
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logger.Debugf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}

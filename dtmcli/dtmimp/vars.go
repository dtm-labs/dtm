/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"errors"

	"github.com/go-resty/resty/v2"
)

// ErrFailure error of FAILURE
var ErrFailure = errors.New("FAILURE")

// ErrOngoing error of ONGOING
var ErrOngoing = errors.New("ONGOING")

// XaSqlTimeoutMs milliseconds for Xa sql to timeout
var XaSqlTimeoutMs = 15000

// MapSuccess HTTP result of SUCCESS
var MapSuccess = map[string]interface{}{"dtm_result": ResultSuccess}

// MapFailure HTTP result of FAILURE
var MapFailure = map[string]interface{}{"dtm_result": ResultFailure}

// RestyClient the resty object
var RestyClient = resty.New()

func init() {
	// RestyClient.SetTimeout(3 * time.Second)
	// RestyClient.SetRetryCount(2)
	// RestyClient.SetRetryWaitTime(1 * time.Second)
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.URL = MayReplaceLocalhost(r.URL)
		Logf("requesting: %s %s %v %v", r.Method, r.URL, r.Body, r.QueryParam)
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		Logf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}

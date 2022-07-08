/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/go-resty/resty/v2"
)

// DB interface
type DB = dtmimp.DB

// TransOptions transaction option
type TransOptions = dtmimp.TransOptions

// DBConf declares db configuration
type DBConf = dtmimp.DBConf

// SetCurrentDBType set currentDBType
func SetCurrentDBType(dbType string) {
	dtmimp.SetCurrentDBType(dbType)
}

// GetCurrentDBType get currentDBType
func GetCurrentDBType() string {
	return dtmimp.GetCurrentDBType()
}

// SetXaSQLTimeoutMs set XaSQLTimeoutMs
func SetXaSQLTimeoutMs(ms int) {
	dtmimp.XaSQLTimeoutMs = ms
}

// GetXaSQLTimeoutMs get XaSQLTimeoutMs
func GetXaSQLTimeoutMs() int {
	return dtmimp.XaSQLTimeoutMs
}

// SetBarrierTableName sets barrier table name
func SetBarrierTableName(tablename string) {
	dtmimp.BarrierTableName = tablename
}

// GetRestyClient get the resty.Client for http request
func GetRestyClient() *resty.Client {
	return dtmimp.RestyClient
}

// SetPassthroughHeaders experimental.
// apply to http header and grpc metadata
// dtm server will save these headers in trans creating request.
// and then passthrough them to sub-trans
func SetPassthroughHeaders(headers []string) {
	dtmimp.PassthroughHeaders = headers
}

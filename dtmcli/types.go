/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/go-resty/resty/v2"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := map[string]string{}
	resp, err := dtmimp.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// DB interface
type DB = dtmimp.DB

// TransOptions transaction option
type TransOptions = dtmimp.TransOptions

// DBConf declares db configuration
type DBConf = dtmimp.DBConf

// String2DtmError translate string to dtm error
func String2DtmError(str string) error {
	return map[string]error{
		ResultFailure: ErrFailure,
		ResultOngoing: ErrOngoing,
		ResultSuccess: nil,
		"":            nil,
	}[str]
}

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

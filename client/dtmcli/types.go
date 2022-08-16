/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
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

// SetBarrierTableName sets barrier table name
func SetBarrierTableName(tablename string) {
	dtmimp.BarrierTableName = tablename
}

// GetRestyClient get the resty.Client for http request
func GetRestyClient() *resty.Client {
	return dtmimp.GetRestyClient2(0)
}

// GetRestyClient2 get the resty.Client with the specified timeout set
func GetRestyClient2(timeout time.Duration) *resty.Client {
	return dtmimp.GetRestyClient2(timeout)
}

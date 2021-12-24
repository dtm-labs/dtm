/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
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

type DBConf = dtmimp.DBConf

// SetCurrentDBType set currentDBType
func SetCurrentDBType(dbType string) {
	dtmimp.SetCurrentDBType(dbType)
}

// GetCurrentDBType get currentDBType
func GetCurrentDBType() string {
	return dtmimp.GetCurrentDBType()
}

// SetXaSqlTimeoutMs set XaSqlTimeoutMs
func SetXaSqlTimeoutMs(ms int) {
	dtmimp.XaSqlTimeoutMs = ms
}

// GetXaSqlTimeoutMs get XaSqlTimeoutMs
func GetXaSqlTimeoutMs() int {
	return dtmimp.XaSqlTimeoutMs
}

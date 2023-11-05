/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
)

const (
	// StatusPrepared status for global/branch trans status.
	// first step, tx preparation period
	StatusPrepared = "prepared"
	// StatusSubmitted status for global trans status.
	StatusSubmitted = "submitted"
	// StatusSucceed status for global/branch trans status.
	StatusSucceed = "succeed"
	// StatusFailed status for global/branch trans status.
	// NOTE: change global status to failed can stop trigger (Not recommended in production env)
	StatusFailed = "failed"
	// StatusAborting status for global trans status.
	StatusAborting = "aborting"

	// ResultSuccess for result of a trans/trans branch
	ResultSuccess = dtmimp.ResultSuccess
	// ResultFailure for result of a trans/trans branch
	ResultFailure = dtmimp.ResultFailure
	// ResultOngoing for result of a trans/trans branch
	ResultOngoing = dtmimp.ResultOngoing

	// DBTypeMysql const for driver mysql
	DBTypeMysql = dtmimp.DBTypeMysql
	// DBTypePostgres const for driver postgres
	DBTypePostgres = dtmimp.DBTypePostgres
	// DBTypeSQLServer const for driver SQLServer
	DBTypeSQLServer = dtmimp.DBTypeSQLServer
)

// MapSuccess HTTP result of SUCCESS
var MapSuccess = dtmimp.MapSuccess

// MapFailure HTTP result of FAILURE
var MapFailure = dtmimp.MapFailure

// ErrFailure error for returned failure
var ErrFailure = dtmimp.ErrFailure

// ErrOngoing error for returned ongoing
var ErrOngoing = dtmimp.ErrOngoing

// ErrDuplicated error of DUPLICATED for only msg
// if QueryPrepared executed before call. then DoAndSubmit return this error
var ErrDuplicated = dtmimp.ErrDuplicated

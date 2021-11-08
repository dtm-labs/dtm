package dtmcli

import (
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

const (
	// StatusPrepared status for global/branch trans status.
	StatusPrepared = "prepared"
	// StatusSubmitted status for global trans status.
	StatusSubmitted = "submitted"
	// StatusSucceed status for global/branch trans status.
	StatusSucceed = "succeed"
	// StatusFailed status for global/branch trans status.
	StatusFailed = "failed"
	// StatusAborting status for global trans status.
	StatusAborting = "aborting"

	// BranchTry branch type for TCC
	BranchTry = "try"
	// BranchConfirm branch type for TCC
	BranchConfirm = "confirm"
	// BranchCancel branch type for TCC
	BranchCancel = "cancel"
	// BranchAction branch type for message, SAGA, XA
	BranchAction = "action"
	// BranchCompensate branch type for SAGA
	BranchCompensate = "compensate"
	// BranchCommit branch type for XA
	BranchCommit = "commit"
	// BranchRollback branch type for XA
	BranchRollback = "rollback"

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
)

// MapSuccess HTTP result of SUCCESS
var MapSuccess = dtmimp.MapSuccess

// MapFailure HTTP result of FAILURE
var MapFailure = dtmimp.MapSuccess

// ErrFailure error for returned failure
var ErrFailure = dtmimp.ErrFailure

// ErrOngoing error for returned ongoing
var ErrOngoing = dtmimp.ErrOngoing

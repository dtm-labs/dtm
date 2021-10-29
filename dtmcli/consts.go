package dtmcli

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
	ResultSuccess = "SUCCESS"
	// ResultFailure for result of a trans/trans branch
	ResultFailure = "FAILURE"
	// ResultOngoing for result of a trans/trans branch
	ResultOngoing = "ONGOING"

	// DBTypeMysql const for driver mysql
	DBTypeMysql = "mysql"
	// DBTypePostgres const for driver postgres
	DBTypePostgres = "postgres"
)

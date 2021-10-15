package dtmcli

const (
	// StatusPrepared status for global trans status. exists only in tran message
	StatusPrepared = "prepared"
	// StatusSubmitted StatusSubmitted status for global trans status.
	StatusSubmitted = "submitted"
	// StatusSucceed status for global trans status.
	StatusSucceed = "succeed"
	// StatusFailed status for global trans status.
	StatusFailed = "failed"

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

	// DBTypeMysql const for driver mysql
	DBTypeMysql = "mysql"
	// DBTypePostgres const for driver postgres
	DBTypePostgres = "postgres"
)

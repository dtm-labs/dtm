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

	// DriverMysql const for driver mysql
	DriverMysql = "mysql"
	// DriverPostgres const for driver postgres
	DriverPostgres = "postgres"
)

// DBDriver dtm和dtmcli可以支持mysql和postgres，但不支持混合，通过全局变量指定当前要支持的驱动
var DBDriver = DriverMysql

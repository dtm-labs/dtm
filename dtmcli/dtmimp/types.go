package dtmimp

import "database/sql"

// DB inteface of dtmcli db
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Tx interface of dtmcli tx
type Tx interface {
	Rollback() error
	Commit() error
	DB
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

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

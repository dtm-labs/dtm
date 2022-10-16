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

// DBConf defines db config
type DBConf struct {
	Driver   string `yaml:"Driver"`
	Host     string `yaml:"Host"`
	Port     int64  `yaml:"Port"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	Db       string `yaml:"Db"`
	Schema   string `yaml:"Schema"`
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"fmt"
	"strings"
)

// DBSpecial db specific operations
type DBSpecial interface {
	GetPlaceHoldSQL(sql string) string
	GetInsertIgnoreTemplate(tableAndValues string, pgConstraint string) string
	GetXaSQL(command string, xid string) string
}

var dbSpecials = map[string]DBSpecial{}
var currentDBType = DBTypeMysql

type mysqlDBSpecial struct{}

func (*mysqlDBSpecial) GetPlaceHoldSQL(sql string) string {
	return sql
}

func (*mysqlDBSpecial) GetXaSQL(command string, xid string) string {
	return fmt.Sprintf("xa %s '%s'", command, xid)
}

func (*mysqlDBSpecial) GetInsertIgnoreTemplate(tableAndValues string, pgConstraint string) string {
	return fmt.Sprintf("insert ignore into %s", tableAndValues)
}

func init() {
	dbSpecials[DBTypeMysql] = &mysqlDBSpecial{}
}

type postgresDBSpecial struct{}

func (*postgresDBSpecial) GetXaSQL(command string, xid string) string {
	return map[string]string{
		"end":      "",
		"start":    "begin",
		"prepare":  fmt.Sprintf("prepare transaction '%s'", xid),
		"commit":   fmt.Sprintf("commit prepared '%s'", xid),
		"rollback": fmt.Sprintf("rollback prepared '%s'", xid),
	}[command]
}

func (*postgresDBSpecial) GetPlaceHoldSQL(sql string) string {
	pos := 1
	parts := []string{}
	b := 0
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' {
			parts = append(parts, sql[b:i])
			b = i + 1
			parts = append(parts, fmt.Sprintf("$%d", pos))
			pos++
		}
	}
	parts = append(parts, sql[b:])
	return strings.Join(parts, "")
}

func (*postgresDBSpecial) GetInsertIgnoreTemplate(tableAndValues string, pgConstraint string) string {
	return fmt.Sprintf("insert into %s on conflict ON CONSTRAINT %s do nothing", tableAndValues, pgConstraint)
}
func init() {
	dbSpecials[DBTypePostgres] = &postgresDBSpecial{}
}

// GetDBSpecial get DBSpecial for currentDBType
func GetDBSpecial(dbType string) DBSpecial {
	if dbType == "" {
		dbType = currentDBType
	}
	return dbSpecials[dbType]
}

// SetCurrentDBType set currentDBType
func SetCurrentDBType(dbType string) {
	spec := dbSpecials[dbType]
	PanicIf(spec == nil, fmt.Errorf("unknown db type '%s'", dbType))
	currentDBType = dbType
}

// GetCurrentDBType get currentDBType
func GetCurrentDBType() string {
	return currentDBType
}

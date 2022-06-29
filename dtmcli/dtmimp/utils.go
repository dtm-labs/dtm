/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/go-resty/resty/v2"
)

// Logf an alias of Infof
// Deprecated: use logger.Errorf
var Logf = logger.Infof

// LogRedf an alias of Errorf
// Deprecated: use logger.Errorf
var LogRedf = logger.Errorf

// FatalIfError fatal if error is not nil
// Deprecated: use logger.FatalIfError
var FatalIfError = logger.FatalIfError

// LogIfFatalf fatal if cond is true
// Deprecated: use logger.FatalfIf
var LogIfFatalf = logger.FatalfIf

// AsError wrap a panic value as an error
func AsError(x interface{}) error {
	logger.Errorf("panic wrapped to error: '%v'", x)
	if e, ok := x.(error); ok {
		return e
	}
	return fmt.Errorf("%v", x)
}

// P2E panic to error
func P2E(perr *error) {
	if x := recover(); x != nil {
		*perr = AsError(x)
	}
}

// E2P error to panic
func E2P(err error) {
	if err != nil {
		panic(err)
	}
}

// CatchP catch panic to error
func CatchP(f func()) (rerr error) {
	defer P2E(&rerr)
	f()
	return nil
}

// PanicIf name is clear
func PanicIf(cond bool, err error) {
	if cond {
		panic(err)
	}
}

// MustAtoi is string to int
func MustAtoi(s string) int {
	r, err := strconv.Atoi(s)
	if err != nil {
		E2P(errors.New("convert to int error: " + s))
	}
	return r
}

// OrString return the first not empty string
func OrString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// If ternary operator
func If(condition bool, trueObj interface{}, falseObj interface{}) interface{} {
	if condition {
		return trueObj
	}
	return falseObj
}

// MustMarshal checked version for marshal
func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	E2P(err)
	return b
}

// MustMarshalString string version of MustMarshal
func MustMarshalString(v interface{}) string {
	return string(MustMarshal(v))
}

// MustUnmarshal checked version for unmarshal
func MustUnmarshal(b []byte, obj interface{}) {
	err := json.Unmarshal(b, obj)
	E2P(err)
}

// MustUnmarshalString string version of MustUnmarshal
func MustUnmarshalString(s string, obj interface{}) {
	MustUnmarshal([]byte(s), obj)
}

// MustRemarshal marshal and unmarshal, and check error
func MustRemarshal(from interface{}, to interface{}) {
	b, err := json.Marshal(from)
	E2P(err)
	err = json.Unmarshal(b, to)
	E2P(err)
}

// GetFuncName get current call func name
func GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	nm := runtime.FuncForPC(pc).Name()
	return nm[strings.LastIndex(nm, ".")+1:]
}

// MayReplaceLocalhost when run in docker compose, change localhost to host.docker.internal for accessing host network
func MayReplaceLocalhost(host string) string {
	if os.Getenv("IS_DOCKER") != "" {
		return strings.Replace(strings.Replace(host,
			"localhost", "host.docker.internal", 1),
			"127.0.0.1", "host.docker.internal", 1)
	}
	return host
}

var sqlDbs sync.Map

// PooledDB get pooled sql.DB
func PooledDB(conf DBConf) (*sql.DB, error) {
	dsn := GetDsn(conf)
	db, ok := sqlDbs.Load(dsn)
	if !ok {
		db2, err := StandaloneDB(conf)
		if err != nil {
			return nil, err
		}
		db = db2
		sqlDbs.Store(dsn, db)
	}
	return db.(*sql.DB), nil
}

// StandaloneDB get a standalone db instance
func StandaloneDB(conf DBConf) (*sql.DB, error) {
	dsn := GetDsn(conf)
	logger.Infof("opening standalone %s: %s", conf.Driver, strings.Replace(dsn, conf.Password, "****", 1))
	return sql.Open(conf.Driver, dsn)
}

// XaDB return a standalone db instance for xa
func XaDB(conf DBConf) (*sql.DB, error) {
	dsn := GetDsn(conf)
	if conf.Driver == DBTypeMysql {
		dsn += "&autocommit=0"
	}
	logger.Infof("opening standalone %s: %s", conf.Driver, strings.Replace(dsn, conf.Password, "****", 1))
	return sql.Open(conf.Driver, dsn)
}

// DBExec use raw db to exec
func DBExec(dbType string, db DB, sql string, values ...interface{}) (affected int64, rerr error) {
	if sql == "" {
		return 0, nil
	}
	began := time.Now()
	sql = GetDBSpecial(dbType).GetPlaceHoldSQL(sql)
	r, rerr := db.Exec(sql, values...)
	used := time.Since(began) / time.Millisecond
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		logger.Debugf("used: %d ms affected: %d for %s %v", used, affected, sql, values)
	} else {
		logger.Errorf("used: %d ms exec error: %v for %s %v", used, rerr, sql, values)
	}
	return
}

// GetDsn get dsn from map config
func GetDsn(conf DBConf) string {
	host := MayReplaceLocalhost(conf.Host)
	driver := conf.Driver
	dsn := map[string]string{
		"mysql": fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&interpolateParams=true",
			conf.User, conf.Password, host, conf.Port, conf.Db),
		"postgres": fmt.Sprintf("host=%s user=%s password=%s dbname='%s' port=%d sslmode=disable",
			host, conf.User, conf.Password, conf.Db, conf.Port),
	}[driver]
	PanicIf(dsn == "", fmt.Errorf("unknow driver: %s", driver))
	return dsn
}

// RespAsErrorCompatible translate a resty response to error
// compatible with version < v1.10
func RespAsErrorCompatible(resp *resty.Response) error {
	code := resp.StatusCode()
	str := resp.String()
	if code == http.StatusTooEarly || strings.Contains(str, ResultOngoing) {
		return fmt.Errorf("%s. %w", str, ErrOngoing)
	} else if code == http.StatusConflict || strings.Contains(str, ResultFailure) {
		return fmt.Errorf("%s. %w", str, ErrFailure)
	} else if code != http.StatusOK {
		return errors.New(str)
	}
	return nil
}

// RespAsErrorByJSONRPC  translate json rpc resty response to error
func RespAsErrorByJSONRPC(resp *resty.Response) error {
	str := resp.String()
	var result map[string]interface{}
	MustUnmarshalString(str, &result)
	if result["error"] != nil {
		rerr := result["error"].(map[string]interface{})
		if rerr["code"] == JrpcCodeFailure {
			return fmt.Errorf("%s. %w", str, ErrFailure)
		} else if rerr["code"] == JrpcCodeOngoing {
			return ErrOngoing
		}
		return errors.New(resp.String())
	}
	return nil
}

// DeferDo a common defer do used in dtmcli/dtmgrpc
func DeferDo(rerr *error, success func() error, fail func() error) {
	if x := recover(); x != nil {
		*rerr = AsError(x)
		_ = fail()
		panic(x)
	} else if *rerr != nil {
		_ = fail()
	} else {
		*rerr = success()
	}
}

// Escape solve CodeQL reported problem
func Escape(input string) string {
	v := strings.Replace(input, "\n", "", -1)
	v = strings.Replace(v, "\r", "", -1)
	v = strings.Replace(v, ";", "", -1)
	// v = strings.Replace(v, "'", "", -1)
	return v
}

// EscapeGet escape get
func EscapeGet(qs url.Values, key string) string {
	return Escape(qs.Get(key))
}

// InsertBarrier insert a record to barrier
func InsertBarrier(tx DB, transType string, gid string, branchID string, op string, barrierID string, reason string, dbType string, barrierTableName string) (int64, error) {
	if op == "" {
		return 0, nil
	}
	if dbType == "" {
		dbType = currentDBType
	}
	if barrierTableName == "" {
		barrierTableName = BarrierTableName
	}
	sql := GetDBSpecial(dbType).GetInsertIgnoreTemplate(barrierTableName+"(trans_type, gid, branch_id, op, barrier_id, reason) values(?,?,?,?,?,?)", "uniq_barrier")
	return DBExec(dbType, tx, sql, transType, gid, branchID, op, barrierID, reason)
}

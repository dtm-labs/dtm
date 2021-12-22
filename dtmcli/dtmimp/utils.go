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
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AsError wrap a panic value as an error
func AsError(x interface{}) error {
	LogRedf("panic wrapped to error: '%v'", x)
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

var logger *zap.SugaredLogger = nil

func init() {
	InitLog()
}

// InitLog is a initialization for a logger
func InitLog() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if os.Getenv("DTM_DEBUG") != "" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	p, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatal("create logger failed: ", err)
	}
	logger = p.Sugar()
}

// Logf is log stdout
func Logf(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// LogRedf is print error message with red color
func LogRedf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
}

// FatalExitFunc is a Fatal exit function ，it will be replaced when testing
var FatalExitFunc = func() { os.Exit(1) }

// LogFatalf is print error message with red color, and execute FatalExitFunc
func LogFatalf(fmt string, args ...interface{}) {
	fmt += "\n" + string(debug.Stack())
	LogRedf(fmt, args...)
	FatalExitFunc()
}

// LogIfFatalf is print error message with red color, and execute LogFatalf, when condition is true
func LogIfFatalf(condition bool, fmt string, args ...interface{}) {
	if condition {
		LogFatalf(fmt, args...)
	}
}

// FatalIfError is print error message with red color, and execute LogIfFatalf.
func FatalIfError(err error) {
	LogIfFatalf(err != nil, "Fatal error: %v", err)
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
		return strings.Replace(host, "localhost", "host.docker.internal", 1)
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
	Logf("opening standalone %s: %s", conf.Driver, strings.Replace(dsn, conf.Passwrod, "****", 1))
	return sql.Open(conf.Driver, dsn)
}

// DBExec use raw db to exec
func DBExec(db DB, sql string, values ...interface{}) (affected int64, rerr error) {
	if sql == "" {
		return 0, nil
	}
	began := time.Now()
	sql = GetDBSpecial().GetPlaceHoldSQL(sql)
	r, rerr := db.Exec(sql, values...)
	used := time.Since(began) / time.Millisecond
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		Logf("used: %d ms affected: %d for %s %v", used, affected, sql, values)
	} else {
		LogRedf("used: %d ms exec error: %v for %s %v", used, rerr, sql, values)
	}
	return
}

// GetDsn get dsn from map config
func GetDsn(conf DBConf) string {
	host := MayReplaceLocalhost(conf.Host)
	driver := conf.Driver
	dsn := map[string]string{
		"mysql": fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			conf.User, conf.Passwrod, host, conf.Port, ""),
		"postgres": fmt.Sprintf("host=%s user=%s password=%s dbname='%s' port=%d sslmode=disable",
			host, conf.User, conf.Passwrod, "", conf.Port),
	}[driver]
	PanicIf(dsn == "", fmt.Errorf("unknow driver: %s", driver))
	return dsn
}

// CheckResponse is check response, and return corresponding error by the condition of resp when err is nil. Otherwise, return err directly.
func CheckResponse(resp *resty.Response, err error) error {
	if err == nil && resp != nil {
		if resp.IsError() {
			return errors.New(resp.String())
		} else if strings.Contains(resp.String(), ResultFailure) {
			return ErrFailure
		} else if strings.Contains(resp.String(), ResultOngoing) {
			return ErrOngoing
		}
	}
	return err
}

// CheckResult is check result. Return err directly if err is not nil. And return corresponding error by calling CheckResponse if resp is the type of *resty.Response. 
// Otherwise, return error by value of str, the string after marshal.
func CheckResult(res interface{}, err error) error {
	if err != nil {
		return err
	}
	resp, ok := res.(*resty.Response)
	if ok {
		return CheckResponse(resp, err)
	}
	if res != nil {
		str := MustMarshalString(res)
		if strings.Contains(str, ResultFailure) {
			return ErrFailure
		} else if strings.Contains(str, ResultOngoing) {
			return ErrOngoing
		}
	}
	return err
}

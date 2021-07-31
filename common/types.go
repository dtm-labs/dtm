package common

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"

	// "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// M a short name
type M = map[string]interface{}

// MS a short name
type MS = map[string]string

// ModelBase model base for gorm to provide base fields
type ModelBase struct {
	ID         uint
	CreateTime *time.Time `gorm:"autoCreateTime"`
	UpdateTime *time.Time `gorm:"autoUpdateTime"`
}

func getGormDialator(driver string, dsn string) gorm.Dialector {
	if driver == "mysql" {
		return mysql.Open(dsn)
		// } else if driver == "postgres" {
		// 	return postgres.Open(dsn)
	}
	panic(fmt.Errorf("unkown driver: %s", driver))
}

var dbs = map[string]*DB{}
var sqlDbs = map[string]*sql.DB{}

// DB provide more func over gorm.DB
type DB struct {
	*gorm.DB
}

// Must set must flag, panic when error occur
func (m *DB) Must() *DB {
	db := m.InstanceSet("ivy.must", true)
	return &DB{DB: db}
}

// NoMust unset must flag, don't panic when error occur
func (m *DB) NoMust() *DB {
	db := m.InstanceSet("ivy.must", false)
	return &DB{DB: db}
}

// ToSQLDB get the sql.DB
func (m *DB) ToSQLDB() *sql.DB {
	d, err := m.DB.DB()
	E2P(err)
	return d
}

type tracePlugin struct{}

func (op *tracePlugin) Name() string {
	return "tracePlugin"
}

func (op *tracePlugin) Initialize(db *gorm.DB) (err error) {
	before := func(db *gorm.DB) {
		db.InstanceSet("ivy.startTime", time.Now())
	}

	after := func(db *gorm.DB) {
		_ts, _ := db.InstanceGet("ivy.startTime")
		sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
		logrus.Printf("used: %d ms affected: %d sql is: %s", time.Since(_ts.(time.Time)).Milliseconds(), db.RowsAffected, sql)
		if v, ok := db.InstanceGet("ivy.must"); ok && v.(bool) {
			if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
				panic(db.Error)
			}
		}
	}

	beforeName := "cb_before"
	afterName := "cb_after"

	logrus.Printf("installing db plugin: %s", op.Name())
	// 开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(beforeName, before)
	_ = db.Callback().Query().Before("gorm:query").Register(beforeName, before)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(beforeName, before)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(beforeName, before)
	_ = db.Callback().Row().Before("gorm:row").Register(beforeName, before)
	_ = db.Callback().Raw().Before("gorm:raw").Register(beforeName, before)

	// 结束后
	_ = db.Callback().Create().After("gorm:after_create").Register(afterName, after)
	_ = db.Callback().Query().After("gorm:after_query").Register(afterName, after)
	_ = db.Callback().Delete().After("gorm:after_delete").Register(afterName, after)
	_ = db.Callback().Update().After("gorm:after_update").Register(afterName, after)
	_ = db.Callback().Row().After("gorm:row").Register(afterName, after)
	_ = db.Callback().Raw().After("gorm:raw").Register(afterName, after)
	return
}

// GetDsn get dsn from map config
func GetDsn(conf map[string]string) string {
	conf["host"] = MayReplaceLocalhost(conf["host"])
	driver := conf["driver"]
	dsn := MS{
		"mysql": fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			conf["user"], conf["password"], conf["host"], conf["port"], conf["database"]),
		"postgres": fmt.Sprintf("host=%s user=%s password=%s dbname='%s' port=%s sslmode=disable TimeZone=Asia/Shanghai",
			conf["host"], conf["user"], conf["password"], conf["database"], conf["port"]),
	}[driver]
	PanicIf(dsn == "", fmt.Errorf("unknow driver: %s", driver))
	return dsn
}

// DbGet get db connection for specified conf
func DbGet(conf map[string]string) *DB {
	dsn := GetDsn(conf)
	if dbs[dsn] == nil {
		logrus.Printf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(getGormDialator(conf["driver"], dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		E2P(err)
		db1.Use(&tracePlugin{})
		dbs[dsn] = &DB{DB: db1}
	}
	return dbs[dsn]
}

// SdbGet get pooled sql.DB
func SdbGet(conf map[string]string) *sql.DB {
	dsn := GetDsn(conf)
	if sqlDbs[dsn] == nil {
		sqlDbs[dsn] = SdbAlone(conf)
	}
	return sqlDbs[dsn]
}

// SdbAlone get a standalone db connection
func SdbAlone(conf map[string]string) *sql.DB {
	dsn := GetDsn(conf)
	logrus.Printf("opening alone %s: %s", conf["driver"], strings.Replace(dsn, conf["password"], "****", 1))
	mdb, err := sql.Open(conf["driver"], dsn)
	E2P(err)
	return mdb
}

// SdbExec use raw db to exec
func SdbExec(db *sql.DB, sql string, values ...interface{}) (affected int64, rerr error) {
	r, rerr := db.Exec(sql, values...)
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		logrus.Printf("affected: %d for %s %v", affected, sql, values)
	} else {
		logrus.Printf("\x1b[31m\nexec error: %v for %s %v\x1b[0m\n", rerr, sql, values)
	}
	return
}

// StxExec use raw tx to exec
func StxExec(tx *sql.Tx, sql string, values ...interface{}) (affected int64, rerr error) {
	r, rerr := tx.Exec(sql, values...)
	if rerr == nil {
		affected, rerr = r.RowsAffected()
		logrus.Printf("affected: %d for %s %v", affected, sql, values)
	} else {
		logrus.Printf("\x1b[31m\nexec error: %v for %s %v\x1b[0m\n", rerr, sql, values)
	}
	return
}

// StxQueryRow use raw tx to query row
func StxQueryRow(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	logrus.Printf("querying: "+query, args...)
	return tx.QueryRow(query, args...)
}

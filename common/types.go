package common

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
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

var dbs = map[string]*DB{}

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
	if IsDockerCompose() {
		conf["host"] = strings.Replace(conf["host"], "localhost", "host.docker.internal", 1)
	}
	// logrus.Printf("is docker: %t IS_DOCKER_COMPOSE: %s and conf host: %s", IsDockerCompose(), os.Getenv("IS_DOCKER_COMPOSE"), conf["host"])
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", conf["user"], conf["password"], conf["host"], conf["port"], conf["database"])
}

// ReplaceDsnPassword replace password for log output
func ReplaceDsnPassword(dsn string) string {
	reg := regexp.MustCompile(`:.*@`)
	return reg.ReplaceAllString(dsn, ":****@")
}

// DbGet get db connection for specified conf
func DbGet(conf map[string]string) *DB {
	dsn := GetDsn(conf)
	if dbs[dsn] == nil {
		logrus.Printf("connecting %s", ReplaceDsnPassword(dsn))
		db1, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		E2P(err)
		db1.Use(&tracePlugin{})
		dbs[dsn] = &DB{DB: db1}
	}
	return dbs[dsn]
}

// SQLDB2DB name is clear
func SQLDB2DB(sdb *sql.DB) *DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sdb,
	}), &gorm.Config{})
	E2P(err)
	db.Use(&tracePlugin{})
	return &DB{DB: db}
}

// MyConn for xa alone connection
type MyConn struct {
	Conn *sql.DB
	Dsn  string
}

// Close name is clear
func (conn *MyConn) Close() {
	logrus.Printf("closing alone mysql: %s", ReplaceDsnPassword(conn.Dsn))
	conn.Conn.Close()
}

// DbAlone get a standalone db connection
func DbAlone(conf map[string]string) (*DB, *MyConn) {
	dsn := GetDsn(conf)
	logrus.Printf("opening alone mysql: %s", ReplaceDsnPassword(dsn))
	mdb, err := sql.Open("mysql", dsn)
	E2P(err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: mdb,
	}), &gorm.Config{})
	E2P(err)
	gormDB.Use(&tracePlugin{})
	return &DB{DB: gormDB}, &MyConn{Conn: mdb, Dsn: dsn}
}

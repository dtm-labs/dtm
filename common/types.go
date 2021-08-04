package common

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yedf/dtm/dtmcli"

	// _ "github.com/lib/pq"

	"gorm.io/driver/mysql"

	// "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	dtmcli.E2P(err)
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
		dtmcli.Logf("used: %d ms affected: %d sql is: %s", time.Since(_ts.(time.Time)).Milliseconds(), db.RowsAffected, sql)
		if v, ok := db.InstanceGet("ivy.must"); ok && v.(bool) {
			if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
				panic(db.Error)
			}
		}
	}

	beforeName := "cb_before"
	afterName := "cb_after"

	dtmcli.Logf("installing db plugin: %s", op.Name())
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

// DbGet get db connection for specified conf
func DbGet(conf map[string]string) *DB {
	dsn := dtmcli.GetDsn(conf)
	if dbs[dsn] == nil {
		dtmcli.Logf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(getGormDialator(conf["driver"], dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		dtmcli.E2P(err)
		db1.Use(&tracePlugin{})
		dbs[dsn] = &DB{DB: db1}
	}
	return dbs[dsn]
}

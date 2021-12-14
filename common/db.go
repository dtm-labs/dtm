package common

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // register mysql driver
	_ "github.com/lib/pq"              // register postgres driver
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ModelBase model base for gorm to provide base fields
type ModelBase struct {
	ID         uint64
	CreateTime *time.Time `gorm:"autoCreateTime"`
	UpdateTime *time.Time `gorm:"autoUpdateTime"`
}

func getGormDialetor(driver string, dsn string) gorm.Dialector {
	if driver == dtmcli.DBTypePostgres {
		return postgres.Open(dsn)
	}
	dtmimp.PanicIf(driver != dtmcli.DBTypeMysql, fmt.Errorf("unkown driver: %s", driver))
	return mysql.Open(dsn)
}

var dbs sync.Map

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
	dtmimp.E2P(err)
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
		dtmimp.Logf("used: %d ms affected: %d sql is: %s", time.Since(_ts.(time.Time)).Milliseconds(), db.RowsAffected, sql)
		if v, ok := db.InstanceGet("ivy.must"); ok && v.(bool) {
			if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
				panic(db.Error)
			}
		}
	}

	beforeName := "cb_before"
	afterName := "cb_after"

	dtmimp.Logf("installing db plugin: %s", op.Name())
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

// SetDBConn set db connection conf
func SetDBConn(db *DB) {
	sqldb, _ := db.DB.DB()
	maxOpenCons, err := strconv.Atoi(DtmConfig.DB["max_open_conns"])
	if err == nil {
		sqldb.SetMaxOpenConns(maxOpenCons)
	}
	maxIdleCons, err := strconv.Atoi(DtmConfig.DB["max_idle_conns"])
	if err == nil {
		sqldb.SetMaxIdleConns(maxIdleCons)
	}
	connMaxLifeTime, err := strconv.ParseInt(DtmConfig.DB["conn_max_life_time"], 10, 64)
	if err == nil {
		sqldb.SetConnMaxLifetime(time.Duration(connMaxLifeTime) * time.Minute)
	}
}

// DbGet get db connection for specified conf
func DbGet(conf map[string]string) *DB {
	dsn := dtmimp.GetDsn(conf)
	db, ok := dbs.Load(dsn)
	if !ok {
		dtmimp.Logf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(getGormDialetor(conf["driver"], dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		dtmimp.E2P(err)
		db1.Use(&tracePlugin{})
		db = &DB{DB: db1}
		SetDBConn(db.(*DB))
		dbs.Store(dsn, db)
	}
	return db.(*DB)
}

// WaitDBUp wait for db to go up
func WaitDBUp() {
	sdb, err := dtmimp.StandaloneDB(DtmConfig.DB)
	dtmimp.FatalIfError(err)
	defer func() {
		sdb.Close()
	}()
	for _, err = dtmimp.DBExec(sdb, "select 1"); err != nil; { // wait for mysql to start
		time.Sleep(3 * time.Second)
		_, err = dtmimp.DBExec(sdb, "select 1")
	}
}

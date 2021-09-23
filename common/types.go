package common

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/yedf/dtm/dtmcli"
)

// ModelBase model base for gorm to provide base fields
type ModelBase struct {
	ID         uint
	CreateTime *time.Time `gorm:"autoCreateTime"`
	UpdateTime *time.Time `gorm:"autoUpdateTime"`
}

func getGormDialetor(driver string, dsn string) gorm.Dialector {
	if driver == "mysql" {
		return mysql.Open(dsn)
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
		db1, err := gorm.Open(getGormDialetor(conf["driver"], dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		dtmcli.E2P(err)
		db1.Use(&tracePlugin{})
		dbs[dsn] = &DB{DB: db1}
	}
	return dbs[dsn]
}

type dtmConfigType struct {
	TransCronInterval int64             `yaml:"TransCronInterval"` // 单位秒 当事务等待这个时间之后，还没有变化，则进行一轮处理，包括prepared中的任务和committed的任务
	DB                map[string]string `yaml:"DB"`
	DisableLocalhost  int64             `yaml:"DisableLocalhost"`
	RetryLimit        int64             `yaml:"RetryLimit"`
}

// DtmConfig 配置
var DtmConfig = dtmConfigType{}

func getIntEnv(key string, defaultV string) int64 {
	return int64(dtmcli.MustAtoi(dtmcli.OrString(os.Getenv(key), defaultV)))
}

func init() {
	if len(os.Args) == 1 {
		return
	}
	DtmConfig.TransCronInterval = getIntEnv("TRANS_CRON_INTERVAL", "10")
	DtmConfig.DB = map[string]string{
		"driver":   dtmcli.OrString(os.Getenv("DB_DRIVER"), "mysql"),
		"host":     os.Getenv("DB_HOST"),
		"port":     dtmcli.OrString(os.Getenv("DB_PORT"), "3306"),
		"user":     os.Getenv("DB_USER"),
		"password": os.Getenv("DB_PASSWORD"),
	}
	DtmConfig.DisableLocalhost = getIntEnv("DISABLE_LOCALHOST", "0")
	DtmConfig.RetryLimit = getIntEnv("RETRY_LIMIT", "2000000000")
	cont := []byte{}
	for d := MustGetwd(); d != "" && d != "/"; d = filepath.Dir(d) {
		cont1, err := ioutil.ReadFile(d + "/conf.yml")
		if err != nil {
			cont1, err = ioutil.ReadFile(d + "/conf.sample.yml")
		}
		if cont1 != nil {
			cont = cont1
			break
		}
	}
	if cont != nil && len(cont) != 0 {
		dtmcli.Logf("cont is: \n%s", string(cont))
		err := yaml.Unmarshal(cont, &DtmConfig)
		dtmcli.FatalIfError(err)
	}
	dtmcli.LogIfFatalf(DtmConfig.DB["driver"] == "" || DtmConfig.DB["user"] == "",
		"dtm配置错误. 请访问 http://dtm.pub 查看部署运维环节. check you env, and conf.yml/conf.sample.yml in current and parent path: %s. config is: \n%v", MustGetwd(), DtmConfig)
}

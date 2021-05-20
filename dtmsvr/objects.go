package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type M = map[string]interface{}

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

var db *gorm.DB = nil

type MyDb struct {
	*gorm.DB
}

func (m *MyDb) Must() *MyDb {
	db := m.InstanceSet("ivy.must", true)
	return &MyDb{DB: db}
}

func (m *MyDb) NoMust() *MyDb {
	db := m.InstanceSet("ivy.must", false)
	return &MyDb{DB: db}
}

func DbGet() *MyDb {
	LoadConfig()
	if db == nil {
		conf := viper.GetStringMapString("mysql")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", conf["user"], conf["password"], conf["host"], conf["port"], conf["database"])
		logrus.Printf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		common.PanicIfError(err)
		db1.Use(&tracePlugin{})
		db = db1
	}
	return &MyDb{DB: db}
}

func writeTransLog(gid string, action string, status string, step int, detail string) {
	db := DbGet()
	if detail == "" {
		detail = "{}"
	}
	dbr := db.Table("test1.a_dtrans_log").Create(M{
		"gid":    gid,
		"action": action,
		"status": status,
		"step":   step,
		"detail": detail,
	})
	common.PanicIfError(dbr.Error)
}

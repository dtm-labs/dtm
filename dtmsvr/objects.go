package dtmsvr

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type M = map[string]interface{}

var db *gorm.DB = nil

func DbGet() *gorm.DB {
	LoadConfig()
	if db == nil {
		conf := viper.GetStringMapString("mysql")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", conf["user"], conf["password"], conf["host"], conf["port"], conf["database"])
		logrus.Printf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		common.PanicIfError(err)
		db = db1.Debug()
	}
	return db
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

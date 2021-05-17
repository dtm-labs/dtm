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

var rabbit *Rabbitmq = nil

func RabbitmqGet() *Rabbitmq {
	LoadConfig()
	if rabbit == nil {
		rabbit = RabbitmqNew(&ServerConfig.Rabbitmq)
	}
	return rabbit
}

var db *gorm.DB = nil

func DbGet() *gorm.DB {
	LoadConfig()
	if db == nil {
		conf := viper.GetStringMapString("mysql")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", conf["user"], conf["password"], conf["host"], conf["port"], conf["database"])
		logrus.Printf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		common.PanicIfError(err)
		db = db1.Debug()
	}
	return db
}

package dtmsvr

import (
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server   string         `json:"server"`
	Rabbitmq RabbitmqConfig `json:"rabbitmq"`
}

var ServerConfig Config = Config{}

func LoadConfig() {
	_, file, _, _ := runtime.Caller(0)
	viper.SetConfigFile(filepath.Dir(file) + "/dtmsvr.yml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&ServerConfig); err != nil {
		panic(err)
	}
	logrus.Printf("config is: %v", ServerConfig)
}

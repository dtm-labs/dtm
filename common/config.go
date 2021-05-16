package common

import (
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

func LoadConfig() {
	_, file, _, _ := runtime.Caller(0)
	viper.SetConfigFile(filepath.Dir(file) + "/../dtm.yml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

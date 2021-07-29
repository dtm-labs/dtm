package examples

import "github.com/yedf/dtm/common"

type exampleConfig struct {
	Mysql map[string]string `yaml:"Mysql"`
}

var config = exampleConfig{}

var dbName = "dtm_busi"

func init() {
	common.InitConfig(common.GetProjectDir(), &config)
	config.Mysql["database"] = dbName
}

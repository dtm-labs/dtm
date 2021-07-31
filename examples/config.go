package examples

import "github.com/yedf/dtm/common"

type exampleConfig struct {
	DB map[string]string `yaml:"DB"`
}

var config = exampleConfig{}

var dbName = "dtm_busi"

func init() {
	common.InitConfig(common.GetProjectDir(), &config)
	config.DB["database"] = dbName
}

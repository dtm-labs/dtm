package examples

import "github.com/yedf/dtm/common"

type exampleConfig struct {
	DB map[string]string `yaml:"DB"`
}

var config = exampleConfig{}

func init() {
	common.InitConfig(common.GetProjectDir(), &config)
}

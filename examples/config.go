package examples

import (
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

type exampleConfig struct {
	DB map[string]string `yaml:"DB"`
}

var config = exampleConfig{}

func init() {
	common.InitConfig(dtmcli.GetProjectDir(), &config)
}

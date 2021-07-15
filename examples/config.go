package examples

type exampleConfig struct {
	Mysql map[string]string `yaml:"Mysql"`
}

var Config = exampleConfig{}

var dbName = "dtm_busi"

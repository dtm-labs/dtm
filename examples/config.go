package examples

type exampleConfig struct {
	Mysql map[string]string `yaml:"Mysql"`
}

var config = exampleConfig{}

var dbName = "dtm_busi"

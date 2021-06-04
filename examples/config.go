package examples

type exampleConfig struct {
	Mysql map[string]string
}

var Config = exampleConfig{}

var dbName = "dtm_busi"

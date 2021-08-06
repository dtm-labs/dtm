package common

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/yedf/dtm/dtmcli"
)

type DtmConfigType struct {
	TransCronInterval int64             `yaml:"TransCronInterval"` // 单位秒 当事务等待这个时间之后，还没有变化，则进行一轮处理，包括prepared中的任务和committed的任务
	DB                map[string]string `yaml:"DB"`
}

// DtmConfig 配置
var dtmConfig = DtmConfigType{}

func initConfig() {
	dtmConfig.TransCronInterval = int64(dtmcli.MustAtoi(dtmcli.OrString(os.Getenv("TRANS_CRON_INTERVAL"), "10")))
	dtmConfig.DB = map[string]string{
		"driver":   dtmcli.OrString(os.Getenv("DB_DRIVER"), "mysql"),
		"host":     os.Getenv("DB_HOST"),
		"port":     dtmcli.OrString(os.Getenv("DB_PORT"), "3306"),
		"user":     os.Getenv("DB_USER"),
		"password": os.Getenv("DB_PASSWORD"),
	}
	var cont []byte
	for d := MustGetwd(); d != "" && d != "/"; d = filepath.Dir(d) {
		cont1, err := ioutil.ReadFile(d + "/conf.yml")
		if err != nil {
			cont1, err = ioutil.ReadFile(d + "/conf.sample.yml")
		}
		if cont1 != nil {
			cont = cont1
			break
		}
	}
	if cont != nil && len(cont) != 0 {
		dtmcli.Logf("cont is: \n%s", string(cont))
		err := yaml.Unmarshal(cont, &dtmConfig)
		dtmcli.FatalIfError(err)
	}
	if len(dtmConfig.DB["driver"]) == 0 || len(dtmConfig.DB["user"]) == 0 {
		dtmcli.Logf("dtm配置错误. 请访问 http://dtm.pub 查看部署运维环节. check you env, and conf.yml/conf.sample.yml in current and parent path: %s. config is: \n%v", MustGetwd(), dtmConfig)
	}
}

func GetDBConfig() DtmConfigType {
	initConfig()
	return dtmConfig
}

package common

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"gopkg.in/yaml.v2"
)

const (
	DtmMetricsPort = 8889
	Mysql = "mysql"
	Redis = "redis"
	BoltDb = "boltdb"
)

// MicroService config type for micro service
type MicroService struct {
	Driver   string `yaml:"Driver" default:"default"`
	Target   string `yaml:"Target"`
	EndPoint string `yaml:"EndPoint"`
}

type Store struct {
	Driver          string `yaml:"Driver" default:"boltdb"`
	Host            string `yaml:"Host"`
	Port            int64  `yaml:"Port"`
	User            string `yaml:"User"`
	Password        string `yaml:"Password"`
	MaxOpenConns    int64  `yaml:"MaxOpenConns" default:"500"`
	MaxIdleConns    int64  `yaml:"MaxIdleConns" default:"500"`
	ConnMaxLifeTime int64  `yaml:"ConnMaxLifeTime" default:"5"`
	DataExpire      int64  `yaml:"DataExpire" default:"604800"` // Trans data will expire in 7 days. only for redis/boltdb.
	RedisPrefix     string `yaml:"RedisPrefix" default:"{}"`    // Redis storage prefix. store data to only one slot in cluster
}

func (s *Store) IsDB() bool {
	return s.Driver == dtmcli.DBTypeMysql || s.Driver == dtmcli.DBTypePostgres
}

func (s *Store) GetDBConf() dtmcli.DBConf {
	return dtmcli.DBConf{
		Driver:   s.Driver,
		Host:     s.Host,
		Port:     s.Port,
		User:     s.User,
		Passwrod: s.Password,
	}
}

type configType struct {
	Store             Store         `yaml:"Store"`
	TransCronInterval int64         `yaml:"TransCronInterval" default:"3"`
	TimeoutToFail     int64         `yaml:"TimeoutToFail" default:"35"`
	RetryInterval     int64         `yaml:"RetryInterval" default:"10"`
	HttpPort          int64         `yaml:"HttpPort" default:"36789"`
	GrpcPort          int64         `yaml:"GrpcPort" default:"36790"`
	MicroService      MicroService  `yaml:"MicroService"`
	UpdateBranchSync  int64         `yaml:"UpdateBranchSync"`
	ExamplesDB        dtmcli.DBConf `yaml:"ExamplesDB"`
}

// Config 配置
var Config = configType{}

func MustLoadConfig() {
	loadFromEnv("", &Config)
	cont := []byte{}
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
	if len(cont) != 0 {
		dtmimp.Logf("config is: \n%s", string(cont))
		err := yaml.UnmarshalStrict(cont, &Config)
		dtmimp.FatalIfError(err)
	}
	err := checkConfig()
	dtmimp.LogIfFatalf(err != nil, `config error: '%v'.
	check you env, and conf.yml/conf.sample.yml in current and parent path: %s.
	please visit http://d.dtm.pub to see the config document.
	loaded config is:
	%v`, err, MustGetwd(), Config)
}

func checkConfig() error {
	if Config.RetryInterval < 10 {
		return errors.New("RetryInterval should not be less than 10")
	}
	if Config.TimeoutToFail < Config.RetryInterval {
		return errors.New("TimeoutToFail should not be less than RetryInterval")
	}
	if Config.Store.Driver == BoltDb {
		return nil
	}
	if Config.Store.Driver == Mysql {
		if Config.Store.Host == "" {
			return errors.New("Db host not valid ")
		}
		if Config.Store.Port == 0 {
			return errors.New("Db port not valid ")
		}
		if Config.Store.User == ""{
			return errors.New("Db user not valid ")
		}
		if Config.Store.Password == ""{
			return errors.New("Db password not valid ")
		}
	}
	return nil
}

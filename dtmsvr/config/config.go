package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"gopkg.in/yaml.v2"
)

const (
	DtmMetricsPort = 8889
	Mysql          = "mysql"
	Redis          = "redis"
	BoltDb         = "boltdb"
)

// MicroService config type for micro service
type MicroService struct {
	Driver   string `yaml:"Driver" default:"default"`
	Target   string `yaml:"Target"`
	EndPoint string `yaml:"EndPoint"`
}

type Store struct {
	Driver             string `yaml:"Driver" default:"boltdb"`
	Host               string `yaml:"Host"`
	Port               int64  `yaml:"Port"`
	User               string `yaml:"User"`
	Password           string `yaml:"Password"`
	MaxOpenConns       int64  `yaml:"MaxOpenConns" default:"500"`
	MaxIdleConns       int64  `yaml:"MaxIdleConns" default:"500"`
	ConnMaxLifeTime    int64  `yaml:"ConnMaxLifeTime" default:"5"`
	DataExpire         int64  `yaml:"DataExpire" default:"604800"` // Trans data will expire in 7 days. only for redis/boltdb.
	RedisPrefix        string `yaml:"RedisPrefix" default:"{a}"`   // Redis storage prefix. store data to only one slot in cluster
	TransGlobalTable   string `yaml:"TransGlobalTable" default:"dtm.trans_global"`
	TransBranchOpTable string `yaml:"BranchTransOpTable" default:"dtm.trans_branch_op"`
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
		Password: s.Password,
	}
}

type configType struct {
	Store             Store        `yaml:"Store"`
	TransCronInterval int64        `yaml:"TransCronInterval" default:"3"`
	TimeoutToFail     int64        `yaml:"TimeoutToFail" default:"35"`
	RetryInterval     int64        `yaml:"RetryInterval" default:"10"`
	HttpPort          int64        `yaml:"HttpPort" default:"36789"`
	GrpcPort          int64        `yaml:"GrpcPort" default:"36790"`
	MicroService      MicroService `yaml:"MicroService"`
	UpdateBranchSync  int64        `yaml:"UpdateBranchSync"`
	LogLevel          string       `yaml:"LogLevel" default:"info"`
}

// Config 配置
var Config = configType{}

// MustLoadConfig load config from env and file
func MustLoadConfig(confFile string) {
	loadFromEnv("", &Config)
	if confFile != "" {
		cont, err := ioutil.ReadFile(confFile)
		logger.FatalIfError(err)
		err = yaml.UnmarshalStrict(cont, &Config)
		logger.FatalIfError(err)
	}
	scont, err := json.MarshalIndent(&Config, "", "  ")
	logger.FatalIfError(err)
	logger.Infof("config file: %s loaded config is: \n%s", confFile, scont)
	err = checkConfig(&Config)
	logger.FatalfIf(err != nil, `config error: '%v'.
	please visit http://d.dtm.pub to see the config document.`, err)
}

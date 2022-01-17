package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"gopkg.in/yaml.v2"
)

const (
	// DtmMetricsPort is metric port
	DtmMetricsPort = 8889
	// Mysql is mysql driver
	Mysql = "mysql"
	// Redis is redis driver
	Redis = "redis"
	// BoltDb is boltdb driver
	BoltDb = "boltdb"
	// Postgres is postgres driver
	Postgres = "postgres"
)

// MicroService config type for micro service
type MicroService struct {
	Driver   string `yaml:"Driver" default:"default"`
	Target   string `yaml:"Target"`
	EndPoint string `yaml:"EndPoint"`
}

// Log config customize log
type Log struct {
	Outputs            string `yaml:"Outputs" default:"stderr"`
	RotationEnable     int64  `yaml:"RotationEnable" default:"0"`
	RotationConfigJSON string `yaml:"RotationConfigJSON" default:"{}"`
}

// Store defines storage relevant info
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
	TransBranchOpTable string `yaml:"TransBranchOpTable" default:"dtm.trans_branch_op"`
}

// IsDB checks config driver is mysql or postgres
func (s *Store) IsDB() bool {
	return s.Driver == dtmcli.DBTypeMysql || s.Driver == dtmcli.DBTypePostgres
}

// GetDBConf returns db conf info
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
	Store                         Store        `yaml:"Store"`
	TransCronInterval             int64        `yaml:"TransCronInterval" default:"3"`
	TimeoutToFail                 int64        `yaml:"TimeoutToFail" default:"35"`
	RetryInterval                 int64        `yaml:"RetryInterval" default:"10"`
	RequestTimeout                int64        `yaml:"RequestTimeout" default:"3"`
	HTTPPort                      int64        `yaml:"HttpPort" default:"36789"`
	GrpcPort                      int64        `yaml:"GrpcPort" default:"36790"`
	MicroService                  MicroService `yaml:"MicroService"`
	UpdateBranchSync              int64        `yaml:"UpdateBranchSync"`
	UpdateBranchAsyncGoroutineNum int64        `yaml:"UpdateBranchAsyncGoroutineNum" default:"1"`
	LogLevel                      string       `yaml:"LogLevel" default:"info"`
	Log                           Log          `yaml:"Log"`
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

package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/logger"
	"gopkg.in/yaml.v3"
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
	// SQLServer is SQL Server driver
	SQLServer = "sqlserver"
)

// MicroService config type for microservice based grpc
type MicroService struct {
	Driver   string `yaml:"Driver" default:"default"`
	Target   string `yaml:"Target"`
	EndPoint string `yaml:"EndPoint"`
}

// HTTPMicroService is the config type for microservice based on http, like springcloud
type HTTPMicroService struct {
	Driver          string `yaml:"Driver" default:"default"`
	RegistryType    string `yaml:"RegistryType" default:""`
	RegistryAddress string `yaml:"RegistryAddress" default:""`
	RegistryOptions string `yaml:"RegistryOptions" default:"{}"`
	Target          string `yaml:"Target"`
	EndPoint        string `yaml:"EndPoint"`
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
	Db                 string `yaml:"Db" default:"dtm"`
	Schema             string `yaml:"Schema" default:"public"`
	MaxOpenConns       int64  `yaml:"MaxOpenConns" default:"500"`
	MaxIdleConns       int64  `yaml:"MaxIdleConns" default:"500"`
	ConnMaxLifeTime    int64  `yaml:"ConnMaxLifeTime" default:"5"`
	DataExpire         int64  `yaml:"DataExpire" default:"604800"`        // Trans data will expire in 7 days. only for redis/boltdb.
	FinishedDataExpire int64  `yaml:"FinishedDataExpire" default:"86400"` // finished Trans data will expire in 1 days. only for redis.
	RedisPrefix        string `yaml:"RedisPrefix" default:"{a}"`          // Redis storage prefix. store data to only one slot in cluster
}

// IsDB checks config driver is mysql or postgres
func (s *Store) IsDB() bool {
	return s.Driver == dtmcli.DBTypeMysql || s.Driver == dtmcli.DBTypePostgres || s.Driver == dtmcli.DBTypeSQLServer
}

// GetDBConf returns db conf info
func (s *Store) GetDBConf() dtmcli.DBConf {
	return dtmcli.DBConf{
		Driver:   s.Driver,
		Host:     s.Host,
		Port:     s.Port,
		User:     s.User,
		Password: s.Password,
		Db:       s.Db,
		Schema:   s.Schema,
	}
}

// Type is the type for the config of dtm server
type Type struct {
	Store                         Store            `yaml:"Store"`
	TransCronInterval             int64            `yaml:"TransCronInterval" default:"3"`
	TimeoutToFail                 int64            `yaml:"TimeoutToFail" default:"35"`
	RetryInterval                 int64            `yaml:"RetryInterval" default:"10"`
	RequestTimeout                int64            `yaml:"RequestTimeout" default:"3"`
	HTTPPort                      int64            `yaml:"HttpPort" default:"36789"`
	GrpcPort                      int64            `yaml:"GrpcPort" default:"36790"`
	JSONRPCPort                   int64            `yaml:"JsonRpcPort" default:"36791"`
	MicroService                  MicroService     `yaml:"MicroService"`
	HTTPMicroService              HTTPMicroService `yaml:"HttpMicroService"`
	UpdateBranchSync              int64            `yaml:"UpdateBranchSync" default:"1"`
	UpdateBranchAsyncGoroutineNum int64            `yaml:"UpdateBranchAsyncGoroutineNum" default:"1"`
	LogLevel                      string           `yaml:"LogLevel" default:"info"`
	Log                           Log              `yaml:"Log"`
	TimeZoneOffset                string           `yaml:"TimeZoneOffset"`
	ConfigUpdateInterval          int64            `yaml:"ConfigUpdateInterval" default:"3"`
	AlertRetryLimit               int64            `yaml:"AlertRetryLimit" default:"3"`
	AlertWebHook                  string           `yaml:"AlertWebHook"`
	AdminBasePath                 string           `yaml:"AdminBasePath"`
}

// Config config
var Config = Type{}

// MustLoadConfig load config from env and file
func MustLoadConfig(confFile string) {
	loadFromEnv("", &Config)
	if confFile != "" {
		cont, err := ioutil.ReadFile(confFile)
		logger.FatalIfError(err)
		err = yaml.Unmarshal(cont, &Config)
		logger.FatalIfError(err)
	}
	scont, err := json.MarshalIndent(&Config, "", "  ")
	logger.FatalIfError(err)
	logger.Infof("config file: %s loaded config is: \n%s", confFile, scont)
	err = checkConfig(&Config)
	logger.FatalfIf(err != nil, `config error: '%v'.
	please visit http://d.dtm.pub to see the config document.`, err)
}

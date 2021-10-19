package common

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // register mysql driver
	_ "github.com/lib/pq"              // register postgres driver
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yedf/dtm/dtmcli"
)

// ModelBase model base for gorm to provide base fields
type ModelBase struct {
	ID         uint
	CreateTime *time.Time `gorm:"autoCreateTime"`
	UpdateTime *time.Time `gorm:"autoUpdateTime"`
}

func getGormDialetor(driver string, dsn string) gorm.Dialector {
	if driver == dtmcli.DBTypePostgres {
		return postgres.Open(dsn)
	}
	dtmcli.PanicIf(driver != dtmcli.DBTypeMysql, fmt.Errorf("unkown driver: %s", driver))
	return mysql.Open(dsn)
}

var dbs sync.Map

// DB provide more func over gorm.DB
type DB struct {
	*gorm.DB
}

// Must set must flag, panic when error occur
func (m *DB) Must() *DB {
	db := m.InstanceSet("ivy.must", true)
	return &DB{DB: db}
}

// NoMust unset must flag, don't panic when error occur
func (m *DB) NoMust() *DB {
	db := m.InstanceSet("ivy.must", false)
	return &DB{DB: db}
}

// ToSQLDB get the sql.DB
func (m *DB) ToSQLDB() *sql.DB {
	d, err := m.DB.DB()
	dtmcli.E2P(err)
	return d
}

type tracePlugin struct{}

func (op *tracePlugin) Name() string {
	return "tracePlugin"
}

func (op *tracePlugin) Initialize(db *gorm.DB) (err error) {
	before := func(db *gorm.DB) {
		db.InstanceSet("ivy.startTime", time.Now())
	}

	after := func(db *gorm.DB) {
		_ts, _ := db.InstanceGet("ivy.startTime")
		sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
		dtmcli.Logf("used: %d ms affected: %d sql is: %s", time.Since(_ts.(time.Time)).Milliseconds(), db.RowsAffected, sql)
		if v, ok := db.InstanceGet("ivy.must"); ok && v.(bool) {
			if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
				panic(db.Error)
			}
		}
	}

	beforeName := "cb_before"
	afterName := "cb_after"

	dtmcli.Logf("installing db plugin: %s", op.Name())
	// 开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(beforeName, before)
	_ = db.Callback().Query().Before("gorm:query").Register(beforeName, before)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(beforeName, before)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(beforeName, before)
	_ = db.Callback().Row().Before("gorm:row").Register(beforeName, before)
	_ = db.Callback().Raw().Before("gorm:raw").Register(beforeName, before)

	// 结束后
	_ = db.Callback().Create().After("gorm:after_create").Register(afterName, after)
	_ = db.Callback().Query().After("gorm:after_query").Register(afterName, after)
	_ = db.Callback().Delete().After("gorm:after_delete").Register(afterName, after)
	_ = db.Callback().Update().After("gorm:after_update").Register(afterName, after)
	_ = db.Callback().Row().After("gorm:row").Register(afterName, after)
	_ = db.Callback().Raw().After("gorm:raw").Register(afterName, after)
	return
}

// DbGet get db connection for specified conf
func DbGet(conf map[string]string) *DB {
	dsn := dtmcli.GetDsn(conf)
	db, ok := dbs.Load(dsn)
	if !ok {
		dtmcli.Logf("connecting %s", strings.Replace(dsn, conf["password"], "****", 1))
		db1, err := gorm.Open(getGormDialetor(conf["driver"], dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		dtmcli.E2P(err)
		db1.Use(&tracePlugin{})
		db = &DB{DB: db1}
		dbs.Store(dsn, db)
	}
	return db.(*DB)
}

type dtmConfigType struct {
	TransCronInterval int64             `yaml:"TransCronInterval"` // 单位秒 当事务等待这个时间之后，还没有变化，则进行一轮处理，包括prepared中的任务和committed的任务
	DB                map[string]string `yaml:"DB"`
	DisableLocalhost  int64             `yaml:"DisableLocalhost"`
	UpdateBranchSync  int64             `yaml:"UpdateBranchSync"`
}

// DtmConfig 配置
var DtmConfig = dtmConfigType{}

func getIntEnv(key string, defaultV string) int64 {
	return int64(dtmcli.MustAtoi(dtmcli.OrString(os.Getenv(key), defaultV)))
}

func init() {
	if len(os.Args) == 1 {
		return
	}
	DtmConfig.TransCronInterval = getIntEnv("TRANS_CRON_INTERVAL", "10")
	DtmConfig.DB = map[string]string{
		"driver":   dtmcli.OrString(os.Getenv("DB_DRIVER"), "mysql"),
		"host":     os.Getenv("DB_HOST"),
		"port":     dtmcli.OrString(os.Getenv("DB_PORT"), "3306"),
		"user":     os.Getenv("DB_USER"),
		"password": os.Getenv("DB_PASSWORD"),
	}
	DtmConfig.DisableLocalhost = getIntEnv("DISABLE_LOCALHOST", "0")
	DtmConfig.UpdateBranchSync = getIntEnv("UPDATE_BRANCH_SYNC", "0")
	cont := []byte{}

	if cont = readConfFromNacos(); cont == nil {
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
	}
	if cont != nil && len(cont) != 0 {
		dtmcli.Logf("cont is: \n%s", string(cont))
		err := yaml.Unmarshal(cont, &DtmConfig)
		dtmcli.FatalIfError(err)
	}
	dtmcli.LogIfFatalf(DtmConfig.DB["driver"] == "" || DtmConfig.DB["user"] == "",
		"dtm配置错误. 请访问 http://dtm.pub 查看部署运维环节. check you env, and conf.yml/conf.sample.yml in current and parent path: %s. config is: \n%v", MustGetwd(), DtmConfig)
}

func readConfFromNacos() []byte {
	var err error
	// nacos
	nacosAddr := os.Getenv("NACOS_SVR")
	if nacosAddr == "" {
		fmt.Println("Nacos address not config, skip init nacos.")
		return nil
	}

	addr := strings.Split(nacosAddr, ":")
	port := 8848
	if len(addr) > 1 {
		port, err = strconv.Atoi(addr[1])
		if err != nil {
			port = 8848
		}
	}

	nacosNamespace := os.Getenv("NACOS_NS")
	sc := []constant.ServerConfig{{
		IpAddr: addr[0],
		Port:   uint64(port),
	}}
	cc := constant.ClientConfig{
		NamespaceId:         nacosNamespace, // 当namespace是public时，此处填空字符串。
		TimeoutMs:           uint64(MustInt(GetStr("NACOS_TIMEOUT-MS", "5000"))),
		NotLoadCacheAtStart: true,
		LogDir:              GetStr("NACOS_LOG-DIR", "log"),
		CacheDir:            GetStr("NACOS_CACHE-DIR", "cache"),
		RotateTime:          GetStr("NACOS_ROTATE-TIME", "1h"),
		MaxAge:              MustInt(GetStr("NACOS_MAX-AGE", "3")),
		LogLevel:            GetStr("NACOS_LOG-LEVEL", "error"),
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})

	if err != nil {
		fmt.Println("[ERROR] Nacos client create fail, use local config only: " + err.Error())
	}

	conf := vo.ConfigParam{
		DataId: GetStr("NACOS_DATA-ID", "dtm"),
		Group:  GetStr("NACOS_GROUP", "DEV"),
	}

	content, err := configClient.GetConfig(conf)
	if err != nil {
		panic(err)
	}

	fmt.Println("read conf from nacos")
	fmt.Println(content)

	return []byte(content)
}

func GetStr(key string, defVal ...string) string {
	ret := os.Getenv(key)
	if ret == "" && len(defVal) > 0 {
		return defVal[0]
	}
	return ret
}

func MustInt(in string) int64 {
	n, _ := strconv.Atoi(in)
	return int64(n)
}

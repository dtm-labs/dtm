package examples

import (
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func RunSqlScript(mysql map[string]string, script string) {
	conf := map[string]string{}
	common.MustRemarshal(mysql, &conf)
	conf["database"] = ""
	db, con := common.DbAlone(conf)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	if err != nil {
		e2p(err)
	}
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" {
			continue
		}
		logrus.Printf("executing: '%s'", s)
		db.Must().Exec(s)
	}
}

func PopulateMysql() {
	common.InitApp(common.GetProjectDir(), &Config)
	Config.Mysql["database"] = dbName
	RunSqlScript(Config.Mysql, common.GetCurrentDir()+"/examples.sql")
}

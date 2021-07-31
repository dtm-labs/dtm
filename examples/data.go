package examples

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// RunSQLScript 1
func RunSQLScript(mysql map[string]string, script string, skipDrop bool) {
	conf := map[string]string{}
	common.MustRemarshal(mysql, &conf)
	conf["database"] = ""
	db, con := common.DbAlone(conf)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	e2p(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || skipDrop && strings.Contains(s, "drop") {
			continue
		}
		logrus.Printf("executing: '%s'", s)
		db.Must().Exec(s)
	}
}

// PopulateDB populate example mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/examples.%s.sql", common.GetCurrentCodeDir(), config.DB["driver"])
	RunSQLScript(config.DB, file, skipDrop)
}

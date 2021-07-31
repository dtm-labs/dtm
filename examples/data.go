package examples

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/yedf/dtm/common"
)

// RunSQLScript 1
func RunSQLScript(conf map[string]string, script string, skipDrop bool) {
	con := common.SdbAlone(conf)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	e2p(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || skipDrop && strings.Contains(s, "drop") {
			continue
		}
		_, err = common.SdbExec(con, s)
		e2p(err)
	}
}

// PopulateDB populate example mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/examples.%s.sql", common.GetCurrentCodeDir(), config.DB["driver"])
	RunSQLScript(config.DB, file, skipDrop)
}

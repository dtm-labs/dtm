package examples

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

var config = common.DtmConfig

// RunSQLScript 1
func RunSQLScript(conf map[string]string, script string, skipDrop bool) {
	con, err := dtmimp.StandaloneDB(conf)
	dtmimp.FatalIfError(err)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	dtmimp.FatalIfError(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || (skipDrop && strings.Contains(s, "drop")) {
			continue
		}
		_, err = dtmimp.DBExec(con, s)
		dtmimp.FatalIfError(err)
	}
}

func resetXaData() {
	if config.DB["driver"] != "mysql" {
		return
	}

	db := dbGet()
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}

// PopulateDB populate example mysql data
func PopulateDB(skipDrop bool) {
	sdb := sdbGet()
	for _, err := dtmimp.DBExec(sdb, "select 1"); err != nil; { // wait for mysql to start
		time.Sleep(3 * time.Second)
		_, err = dtmimp.DBExec(sdb, "select 1")
	}

	resetXaData()
	file := fmt.Sprintf("%s/examples.%s.sql", common.GetCallerCodeDir(), config.DB["driver"])
	RunSQLScript(config.DB, file, skipDrop)
	file = fmt.Sprintf("%s/../dtmcli/barrier.%s.sql", common.GetCallerCodeDir(), config.DB["driver"])
	RunSQLScript(config.DB, file, skipDrop)
}

type sampleInfo struct {
	Arg    string
	Action func() string
	Desc   string
}

// Samples 所有的示例都会注册到这里
var Samples = map[string]*sampleInfo{}

func addSample(name string, fn func() string) {
	dtmimp.LogIfFatalf(Samples[name] != nil, "%s already exists", name)
	Samples[name] = &sampleInfo{Arg: name, Action: fn}
}

package examples

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

var config = common.DtmConfig

// RunSQLScript 1
func RunSQLScript(conf map[string]string, script string, skipDrop bool) {
	con, err := dtmcli.StandaloneDB(conf)
	dtmcli.FatalIfError(err)
	defer func() { con.Close() }()
	content, err := ioutil.ReadFile(script)
	dtmcli.FatalIfError(err)
	sqls := strings.Split(string(content), ";")
	for _, sql := range sqls {
		s := strings.TrimSpace(sql)
		if s == "" || (skipDrop && strings.Contains(s, "drop")) {
			continue
		}
		_, err = dtmcli.DBExec(con, s)
		dtmcli.FatalIfError(err)
	}
}

// PopulateDB populate example mysql data
func PopulateDB(skipDrop bool) {
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
	dtmcli.LogIfFatalf(Samples[name] != nil, "%s already exists", name)
	Samples[name] = &sampleInfo{Arg: name, Action: fn}
}

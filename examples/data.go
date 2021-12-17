/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"fmt"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

var config = &common.Config

func resetXaData() {
	if config.ExamplesDB.Driver != "mysql" {
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
	resetXaData()
	file := fmt.Sprintf("%s/examples.%s.sql", common.GetCallerCodeDir(), config.ExamplesDB.Driver)
	common.RunSQLScript(config.ExamplesDB, file, skipDrop)
	file = fmt.Sprintf("%s/../dtmcli/barrier.%s.sql", common.GetCallerCodeDir(), config.ExamplesDB.Driver)
	common.RunSQLScript(config.ExamplesDB, file, skipDrop)
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

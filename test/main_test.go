package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

func TestMain(m *testing.M) {
	dtmsvr.TransProcessedTestChan = make(chan string, 1)
	dtmsvr.CronForwardDuration = 60 * time.Second
	common.DtmConfig.UpdateBranchSync = 1
	dtmsvr.PopulateDB(false)
	examples.PopulateDB(false)
	// 启动组件
	go dtmsvr.StartSvr()
	examples.GrpcStartup()
	app = examples.BaseAppStartup()

	resetXaData()
	m.Run()
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

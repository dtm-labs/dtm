package dtmsvr

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

var (
	dtmsvrPort = 8080
	srv        *http.Server
)

// StartSvr StartSvr
func StartSvr(port int) {
	dtmcli.Logf("start dtmsvr")
	if port == 0 {
		port = dtmsvrPort
	}
	app := common.GetGinApp()
	addRoute(app)
	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			dtmcli.Logf("dtmsvr listen at: %d,err:%s", port, err)
		}
	}()
}

func StopSvr() {
	stopCron()
	dtmcli.Logf("shutdown dtmsvr ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		dtmcli.LogFatalf("dtmsvr shutdown:", err)
	}
	dtmcli.Logf("dtmsvr exiting")
}

// PopulateDB setup mysql data
func PopulateDB(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.%s.sql", common.GetCurrentCodeDir(), config.DB["driver"])
	examples.RunSQLScript(config.DB, file, skipDrop)
}

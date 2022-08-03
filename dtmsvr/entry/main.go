package entry

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/automaxprocs/maxprocs"
)

func usage() {
	cmd := filepath.Base(os.Args[0])
	s := "Usage: %s [options]\n\n"
	fmt.Fprintf(os.Stderr, s, cmd)
	flag.PrintDefaults()
}

var isVersion = flag.Bool("v", false, "Show the version of dtm.")
var isDebug = flag.Bool("d", false, "Set log level to debug.")
var isHelp = flag.Bool("h", false, "Show the help information about dtm.")
var isReset = flag.Bool("r", false, "Reset dtm server data.")
var confFile = flag.String("c", "", "Path to the server configuration file.")

// Main is the entry point of dtm server.
func Main(version *string) (*gin.Engine, *config.Type) {
	flag.Parse()
	if *version == "" {
		*version = "v0.0.0-dev"
	}
	dtmsvr.Version = *version
	if flag.NArg() > 0 || *isHelp {
		usage()
		return nil, nil
	} else if *isVersion {
		fmt.Printf("dtm version: %s\n", *version)
		return nil, nil
	}
	logger.Infof("dtm version is: %s", *version)
	config.MustLoadConfig(*confFile)
	if config.Config.TimeZoneOffset != "" {
		time.Local = time.FixedZone("UTC", dtmimp.MustAtoi(config.Config.TimeZoneOffset)*3600)
	}
	conf := &config.Config
	if *isDebug {
		conf.LogLevel = "debug"
	}
	logger.InitLog2(conf.LogLevel, conf.Log.Outputs, conf.Log.RotationEnable, conf.Log.RotationConfigJSON)
	if *isReset {
		dtmsvr.PopulateDB(false)
	}
	_, _ = maxprocs.Set(maxprocs.Logger(logger.Infof))
	registry.WaitStoreUp()
	app := dtmsvr.StartSvr()       // start dtmsvr api
	go dtmsvr.CronExpiredTrans(-1) // start dtmsvr cron job
	return app, &config.Config
}

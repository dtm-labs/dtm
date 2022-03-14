/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"

	// load the microserver driver
	_ "github.com/dtm-labs/dtmdriver-gozero"
	_ "github.com/dtm-labs/dtmdriver-polaris"
	_ "github.com/dtm-labs/dtmdriver-protocol1"
)

// Version declares version info
var Version string

func version() {
	if Version == "" {
		Version = "0.0.0-dev"
	}
	fmt.Printf("dtm version: %s\n", Version)
}

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

func main() {
	flag.Parse()
	if flag.NArg() > 0 || *isHelp {
		usage()
		return
	} else if *isVersion {
		version()
		return
	}
	logger.Infof("dtm version is: %s", Version)
	config.MustLoadConfig(*confFile)
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
	dtmsvr.StartSvr()              // start dtmsvr api
	go dtmsvr.CronExpiredTrans(-1) // start dtmsvr cron job
	select {}
}

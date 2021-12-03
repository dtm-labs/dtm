/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package common

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/yedf/dtm/dtmcli/dtmimp"
)

const (
	DtmHttpPort    = 36789
	DtmGrpcPort    = 36790
	DtmMetricsPort = 8889
)

// MicroService config type for micro service
type MicroService struct {
	Driver   string `yaml:"Driver"`
	Target   string `yaml:"Target"`
	EndPoint string `yaml:"EndPoint"`
}

type dtmConfigType struct {
	TransCronInterval int64             `yaml:"TransCronInterval"`
	TimeoutToFail     int64             `yaml:"TimeoutToFail"`
	RetryInterval     int64             `yaml:"RetryInterval"`
	DB                map[string]string `yaml:"DB"`
	MicroService      MicroService      `yaml:"MicroService"`
	DisableLocalhost  int64             `yaml:"DisableLocalhost"`
	UpdateBranchSync  int64             `yaml:"UpdateBranchSync"`
}

// DtmConfig 配置
var DtmConfig = dtmConfigType{}

func getIntEnv(key string, defaultV string) int64 {
	return int64(dtmimp.MustAtoi(dtmimp.OrString(os.Getenv(key), defaultV)))
}

func MustLoadConfig() {
	DtmConfig.TransCronInterval = getIntEnv("TRANS_CRON_INTERVAL", "3")
	DtmConfig.TimeoutToFail = getIntEnv("TIMEOUT_TO_FAIL", "35")
	DtmConfig.RetryInterval = getIntEnv("RETRY_INTERVAL", "10")
	DtmConfig.DB = map[string]string{
		"driver":             dtmimp.OrString(os.Getenv("DB_DRIVER"), "mysql"),
		"host":               os.Getenv("DB_HOST"),
		"port":               dtmimp.OrString(os.Getenv("DB_PORT"), "3306"),
		"user":               os.Getenv("DB_USER"),
		"password":           os.Getenv("DB_PASSWORD"),
		"max_open_conns":     dtmimp.OrString(os.Getenv("DB_MAX_OPEN_CONNS"), "500"),
		"max_idle_conns":     dtmimp.OrString(os.Getenv("DB_MAX_IDLE_CONNS"), "500"),
		"conn_max_life_time": dtmimp.OrString(os.Getenv("DB_CONN_MAX_LIFE_TIME"), "5"),
	}
	DtmConfig.MicroService.Driver = dtmimp.OrString(os.Getenv("MICRO_SERVICE_DRIVER"), "default")
	DtmConfig.MicroService.Target = os.Getenv("MICRO_SERVICE_TARGET")
	DtmConfig.MicroService.EndPoint = os.Getenv("MICRO_SERVICE_ENDPOINT")
	DtmConfig.DisableLocalhost = getIntEnv("DISABLE_LOCALHOST", "0")
	DtmConfig.UpdateBranchSync = getIntEnv("UPDATE_BRANCH_SYNC", "0")
	cont := []byte{}
	for d := MustGetwd(); d != "" && d != "/"; d = filepath.Dir(d) {
		cont1, err := ioutil.ReadFile(d + "/conf.yml")
		if err != nil {
			cont1, err = ioutil.ReadFile(d + "/conf.sample.yml")
		}
		if cont1 != nil {
			cont = cont1
			break
		}
	}
	if len(cont) != 0 {
		dtmimp.Logf("config is: \n%s", string(cont))
		err := yaml.Unmarshal(cont, &DtmConfig)
		dtmimp.FatalIfError(err)
	}
	err := checkConfig()
	dtmimp.LogIfFatalf(err != nil, `config error: '%v'.
	check you env, and conf.yml/conf.sample.yml in current and parent path: %s.
	please visit http://d.dtm.pub to see the config document.
	loaded config is:
	%v`, err, MustGetwd(), DtmConfig)
}

func checkConfig() error {
	if DtmConfig.DB["driver"] == "" {
		return errors.New("db driver empty")
	} else if DtmConfig.DB["user"] == "" || DtmConfig.DB["host"] == "" {
		return errors.New("db config not valid")
	} else if DtmConfig.RetryInterval < 10 {
		return errors.New("RetryInterval should not be less than 10")
	} else if DtmConfig.TimeoutToFail < DtmConfig.RetryInterval {
		return errors.New("TimeoutToFail should not be less than RetryInterval")
	}
	return nil
}

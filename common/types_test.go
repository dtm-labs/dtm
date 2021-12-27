/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package common

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/stretchr/testify/assert"
)

func TestGeneralDB(t *testing.T) {
	MustLoadConfig()
	if Config.Store.IsDB() {
		testSql(t)
		testDbAlone(t)
	}
}

func testSql(t *testing.T) {
	db := DbGet(Config.Store.GetDBConf())
	err := func() (rerr error) {
		defer dtmimp.P2E(&rerr)
		db.Must().Exec("select a")
		return nil
	}()
	assert.NotEqual(t, nil, err)
}

func testDbAlone(t *testing.T) {
	db, err := dtmimp.StandaloneDB(Config.Store.GetDBConf())
	assert.Nil(t, err)
	_, err = dtmimp.DBExec(db, "select 1")
	assert.Equal(t, nil, err)
	_, err = dtmimp.DBExec(db, "")
	assert.Equal(t, nil, err)
	db.Close()
	_, err = dtmimp.DBExec(db, "select 1")
	assert.NotEqual(t, nil, err)
}

func TestConfig(t *testing.T) {
	testConfigStringField(&Config.Store.Driver, "", t)
	testConfigStringField(&Config.Store.User, "", t)
	testConfigIntField(&Config.RetryInterval, 9, t)
	testConfigIntField(&Config.TimeoutToFail, 9, t)
}

func testConfigStringField(fd *string, val string, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	*fd = old
}

func testConfigIntField(fd *int64, val int64, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	*fd = old
}

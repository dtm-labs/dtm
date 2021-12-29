package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/stretchr/testify/assert"
)

func TestGeneralDB(t *testing.T) {
	if conf.Store.IsDB() {
		testSql(t)
		testDbAlone(t)
	}
}

func testSql(t *testing.T) {
	db := dtmutil.DbGet(conf.Store.GetDBConf())
	err := func() (rerr error) {
		defer dtmimp.P2E(&rerr)
		db.Must().Exec("select a")
		return nil
	}()
	assert.NotEqual(t, nil, err)
}

func testDbAlone(t *testing.T) {
	db, err := dtmimp.StandaloneDB(conf.Store.GetDBConf())
	assert.Nil(t, err)
	_, err = dtmimp.DBExec(db, "select 1")
	assert.Equal(t, nil, err)
	_, err = dtmimp.DBExec(db, "")
	assert.Equal(t, nil, err)
	db.Close()
	_, err = dtmimp.DBExec(db, "select 1")
	assert.NotEqual(t, nil, err)
}

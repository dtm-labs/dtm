package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
)

func TestDb(t *testing.T) {
	db := DbGet(DtmConfig.DB)
	err := func() (rerr error) {
		defer dtmcli.P2E(&rerr)
		dbr := db.NoMust().Exec("select a")
		assert.NotEqual(t, nil, dbr.Error)
		db.Must().Exec("select a")
		return nil
	}()
	assert.NotEqual(t, nil, err)
}

func TestDbAlone(t *testing.T) {
	db, err := dtmcli.StandaloneDB(DtmConfig.DB)
	assert.Nil(t, err)
	_, err = dtmcli.DBExec(db, "select 1")
	assert.Equal(t, nil, err)
	_, err = dtmcli.DBExec(db, "")
	assert.Equal(t, nil, err)
	db.Close()
	_, err = dtmcli.DBExec(db, "select 1")
	assert.NotEqual(t, nil, err)
}

func TestConfig(t *testing.T) {
	testConfigStringField(DtmConfig.DB, "driver", "", t)
	testConfigStringField(DtmConfig.DB, "user", "", t)
	testConfigIntField(&DtmConfig.RetryInterval, 9, t)
	testConfigIntField(&DtmConfig.TimeoutToFail, 9, t)
}

func testConfigStringField(m map[string]string, key string, val string, t *testing.T) {
	old := m[key]
	m[key] = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	m[key] = old
}

func testConfigIntField(fd *int64, val int64, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	*fd = old
}

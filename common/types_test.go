package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
)

type testConfig struct {
	DB map[string]string `yaml:"DB"`
}

var config = testConfig{}

func init() {
	InitConfig(&config)
}

func TestDb(t *testing.T) {
	db := DbGet(config.DB)
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
	db, err := dtmcli.SdbAlone(config.DB)
	assert.Nil(t, err)
	_, err = dtmcli.SdbExec(db, "select 1")
	assert.Equal(t, nil, err)
	db.Close()
	_, err = dtmcli.SdbExec(db, "select 1")
	assert.NotEqual(t, nil, err)
}

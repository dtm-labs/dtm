package common

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

type testConfig struct {
	Mysql map[string]string `yaml:"Mysql"`
}

var config = testConfig{}

var myinit int = func() int {
	InitApp(GetProjectDir(), &config)
	config.Mysql["database"] = ""
	return 0
}()

func TestDb(t *testing.T) {
	db := DbGet(config.Mysql)
	err := func() (rerr error) {
		defer P2E(&rerr)
		dbr := db.NoMust().Exec("select a")
		assert.NotEqual(t, nil, dbr.Error)
		db.Must().Exec("select a")
		return nil
	}()
	assert.NotEqual(t, nil, err)
	sdb := db.ToSQLDB()
	db = SQLDB2DB(sdb)
}

func TestDbAlone(t *testing.T) {
	db, con := DbAlone(config.Mysql)
	dbr := db.Exec("select 1")
	assert.Equal(t, nil, dbr.Error)
	con.Close()
	dbr = db.Exec("select 1")
	assert.NotEqual(t, nil, dbr.Error)
}

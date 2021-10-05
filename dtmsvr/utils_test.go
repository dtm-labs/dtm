package dtmsvr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
)

func TestUtils(t *testing.T) {
	db := dbGet()
	db.NoMust()
	err := dtmcli.CatchP(func() {
		checkAffected(db.DB)
	})
	assert.Error(t, err)

	CronExpiredTrans(1)
	sleepCronTime(10)
}

func TestCheckLocalHost(t *testing.T) {
	config.DisableLocalhost = 1
	err := dtmcli.CatchP(func() {
		checkLocalhost([]TransBranch{{URL: "http://localhost"}})
	})
	assert.Error(t, err)
	config.DisableLocalhost = 0
	err = dtmcli.CatchP(func() {
		checkLocalhost([]TransBranch{{URL: "http://localhost"}})
	})
	assert.Nil(t, err)
}

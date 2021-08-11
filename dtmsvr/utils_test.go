package dtmsvr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
)

func TestUtils(t *testing.T) {
	db := dbGet()
	db.NoMust()
	CronTransOnce(0)
	err := dtmcli.CatchP(func() {
		checkAffected(db.DB)
	})
	assert.Error(t, err)

	CronExpiredTrans(1)
	go sleepCronTime()
}

package dtmimp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBSpecial(t *testing.T) {
	old := currentDBType
	assert.Error(t, CatchP(func() {
		SetCurrentDBType("no-driver")
	}))
	SetCurrentDBType(DBTypeMysql)
	sp := GetDBSpecial()

	assert.Equal(t, "? ?", sp.GetPlaceHoldSQL("? ?"))
	assert.Equal(t, "xa start 'xa1'", sp.GetXaSQL("start", "xa1"))
	assert.Equal(t, "date_add(now(), interval 1000 second)", sp.TimestampAdd(1000))
	SetCurrentDBType(DBTypePostgres)
	sp = GetDBSpecial()
	assert.Equal(t, "$1 $2", sp.GetPlaceHoldSQL("? ?"))
	assert.Equal(t, "begin", sp.GetXaSQL("start", "xa1"))
	assert.Equal(t, "current_timestamp + interval '1000 second'", sp.TimestampAdd(1000))
	SetCurrentDBType(old)
}

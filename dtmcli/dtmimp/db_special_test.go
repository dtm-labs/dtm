/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

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
	sp := GetDBSpecial(DBTypeMysql)

	assert.Equal(t, "? ?", sp.GetPlaceHoldSQL("? ?"))
	assert.Equal(t, "xa start 'xa1'", sp.GetXaSQL("start", "xa1"))
	assert.Equal(t, "insert ignore into a(f) values(?)", sp.GetInsertIgnoreTemplate("a(f) values(?)", "c"))
	SetCurrentDBType(DBTypePostgres)
	sp = GetDBSpecial(DBTypePostgres)
	assert.Equal(t, "$1 $2", sp.GetPlaceHoldSQL("? ?"))
	assert.Equal(t, "begin", sp.GetXaSQL("start", "xa1"))
	assert.Equal(t, "insert into a(f) values(?) on conflict ON CONSTRAINT c do nothing", sp.GetInsertIgnoreTemplate("a(f) values(?)", "c"))
	SetCurrentDBType(old)
}

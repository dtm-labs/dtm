/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmcli

import (
	"net/url"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	err := dtmimp.CatchP(func() {
		MustGenGid("http://localhost:36789/api/no")
	})
	assert.Error(t, err)
	assert.Error(t, err)
	_, err = BarrierFromQuery(url.Values{})
	assert.Error(t, err)

}

func TestXaSqlTimeout(t *testing.T) {
	old := GetXaSQLTimeoutMs()
	SetXaSQLTimeoutMs(old)
	SetBarrierTableName(dtmimp.BarrierTableName) // just cover this func
}

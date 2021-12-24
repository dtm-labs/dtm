/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	"context"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	_, err := BarrierFromGrpc(context.Background())
	assert.Error(t, err)

	_, err = TccFromGrpc(context.Background())
	assert.Error(t, err)

	old := GetCurrentDBType()
	SetCurrentDBType(dtmcli.DBTypeMysql)
	SetCurrentDBType(old)

	err = UseDriver("default")
	assert.Nil(t, err)
}

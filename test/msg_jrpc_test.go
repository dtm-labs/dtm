/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/stretchr/testify/assert"
)

func TestMsgJrpcNormal(t *testing.T) {
	resp, err := dtmcli.GetRestyClient().R().SetBody(map[string]string{
		"jsonrpc": "2.0",
		"method":  "dtmserver.newGid",
		"params":  "",
		"id":      "TestMsgJrpcNormal",
	}).Post(dtmutil.DefaultJrpcServer)
	assert.Nil(t, err)
	assert.Contains(t, resp.String(), "gid")
}

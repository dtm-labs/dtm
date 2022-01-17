/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"fmt"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/stretchr/testify/assert"
)

func TestAPIQuery(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := genMsg(gid).Submit()
	assert.Nil(t, err)
	waitTransProcessed(gid)
	resp, err := dtmimp.RestyClient.R().SetQueryParam("gid", gid).Get(dtmutil.DefaultHTTPServer + "/query")
	assert.Nil(t, err)
	m := map[string]interface{}{}
	assert.Equal(t, resp.StatusCode(), 200)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.NotEqual(t, nil, m["transaction"])
	assert.Equal(t, 2, len(m["branches"].([]interface{})))

	resp, err = dtmimp.RestyClient.R().SetQueryParam("gid", "").Get(dtmutil.DefaultHTTPServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 500)

	resp, err = dtmimp.RestyClient.R().SetQueryParam("gid", "1").Get(dtmutil.DefaultHTTPServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 200)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, nil, m["transaction"])
	assert.Equal(t, 0, len(m["branches"].([]interface{})))
}

func TestAPIAll(t *testing.T) {
	for i := 0; i < 3; i++ { // add three
		gid := dtmimp.GetFuncName() + fmt.Sprintf("%d", i)
		err := genMsg(gid).Submit()
		assert.Nil(t, err)
		waitTransProcessed(gid)
	}
	resp, err := dtmimp.RestyClient.R().SetQueryParam("limit", "1").Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	m := map[string]interface{}{}
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos)

	resp, err = dtmimp.RestyClient.R().SetQueryParams(map[string]string{
		"limit":    "1",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos2 := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos2)
	assert.NotEqual(t, nextPos, nextPos2)

	resp, err = dtmimp.RestyClient.R().SetQueryParams(map[string]string{
		"limit":    "1000",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos3 := m["next_position"].(string)
	assert.Equal(t, "", nextPos3)
}

func TestDtmMetrics(t *testing.T) {
	rest, err := dtmimp.RestyClient.R().Get("http://localhost:36789/api/metrics")
	assert.Nil(t, err)
	assert.Equal(t, rest.StatusCode(), 200)
}

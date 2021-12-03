/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

const gidTestAPI = "TestAPI"

func TestAPIQuery(t *testing.T) {
	err := genMsg(gidTestAPI).Submit()
	assert.Nil(t, err)
	waitTransProcessed(gidTestAPI)
	resp, err := dtmimp.RestyClient.R().SetQueryParam("gid", gidTestAPI).Get(examples.DtmHttpServer + "/query")
	e2p(err)
	m := map[string]interface{}{}
	assert.Equal(t, resp.StatusCode(), 200)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.NotEqual(t, nil, m["transaction"])
	assert.Equal(t, 2, len(m["branches"].([]interface{})))

	resp, err = dtmimp.RestyClient.R().SetQueryParam("gid", "").Get(examples.DtmHttpServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 500)

	resp, err = dtmimp.RestyClient.R().SetQueryParam("gid", "1").Get(examples.DtmHttpServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 200)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, nil, m["transaction"])
	assert.Equal(t, 0, len(m["branches"].([]interface{})))
}

func TestAPIAll(t *testing.T) {
	_, err := dtmimp.RestyClient.R().Get(examples.DtmHttpServer + "/all")
	assert.Nil(t, err)
	_, err = dtmimp.RestyClient.R().SetQueryParam("last_id", "10").Get(examples.DtmHttpServer + "/all")
	assert.Nil(t, err)
	resp, err := dtmimp.RestyClient.R().SetQueryParam("last_id", "abc").Get(examples.DtmHttpServer + "/all")
	assert.Equal(t, resp.StatusCode(), 500)
}

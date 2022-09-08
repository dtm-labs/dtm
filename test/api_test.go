/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestAPIVersion(t *testing.T) {
	resp, err := dtmcli.GetRestyClient().R().Get(dtmutil.DefaultHTTPServer + "/version")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode())
}

func TestAPIQuery(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := genMsg(gid).Submit()
	assert.Nil(t, err)
	waitTransProcessed(gid)
	resp, err := dtmcli.GetRestyClient().R().SetQueryParam("gid", gid).Get(dtmutil.DefaultHTTPServer + "/query")
	assert.Nil(t, err)
	m := map[string]interface{}{}
	assert.Equal(t, resp.StatusCode(), 200)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.NotEqual(t, nil, m["transaction"])
	assert.Equal(t, 2, len(m["branches"].([]interface{})))

	resp, err = dtmcli.GetRestyClient().R().SetQueryParam("gid", "").Get(dtmutil.DefaultHTTPServer + "/query")
	e2p(err)
	assert.Equal(t, resp.StatusCode(), 500)

	resp, err = dtmcli.GetRestyClient().R().SetQueryParam("gid", "1").Get(dtmutil.DefaultHTTPServer + "/query")
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
	resp, err := dtmcli.GetRestyClient().R().SetQueryParam("limit", "1").Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	m := map[string]interface{}{}
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos)

	resp, err = dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"limit":    "1",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos2 := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos2)
	assert.NotEqual(t, nextPos, nextPos2)

	resp, err = dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"limit":    "1000",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/all")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos3 := m["next_position"].(string)
	assert.Equal(t, "", nextPos3)

	//fmt.Printf("pos1:%s,pos2:%s,pos3:%s", nextPos, nextPos2, nextPos3)
}

func TestAPIScanKV(t *testing.T) {
	for i := 0; i < 3; i++ { // add three
		assert.Nil(t, httpSubscribe("test_topic"+fmt.Sprintf("%d", i), "http://dtm/test1"))
	}
	resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat":   "topics",
		"limit": "1",
	}).Get(dtmutil.DefaultHTTPServer + "/scanKV")
	assert.Nil(t, err)
	m := map[string]interface{}{}
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos)

	resp, err = dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat":      "topics",
		"limit":    "1",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/scanKV")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos2 := m["next_position"].(string)
	assert.NotEqual(t, "", nextPos2)
	assert.NotEqual(t, nextPos, nextPos2)

	resp, err = dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat":      "topics",
		"limit":    "1000",
		"position": nextPos,
	}).Get(dtmutil.DefaultHTTPServer + "/scanKV")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	nextPos3 := m["next_position"].(string)
	assert.Equal(t, "", nextPos3)
}

func TestAPIQueryKV(t *testing.T) {
	m := map[string]interface{}{}
	// normal
	assert.Nil(t, httpSubscribe("test_topic_TestAPIQueryKV", "http://dtm/test1"))
	resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat": "topics",
		"key": "test_topic_TestAPIQueryKV",
	}).Get(dtmutil.DefaultHTTPServer + "/queryKV")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, 1, len(m["kv"].([]interface{})))

	// query non_existent topic
	resp, err = dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat": "topics",
		"key": "non_existent_topic_TestAPIQueryKV",
	}).Get(dtmutil.DefaultHTTPServer + "/queryKV")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	assert.Equal(t, 0, len(m["kv"].([]interface{})))
}

func TestDtmMetrics(t *testing.T) {
	rest, err := dtmcli.GetRestyClient().R().Get("http://localhost:36789/api/metrics")
	assert.Nil(t, err)
	assert.Equal(t, rest.StatusCode(), 200)
}

func TestAPIResetCronTime(t *testing.T) {
	testStoreResetCronTime(t, dtmimp.GetFuncName(), func(timeout int64, limit int64) (int64, bool, error) {
		sTimeout := strconv.FormatInt(timeout, 10)
		sLimit := strconv.FormatInt(limit, 10)

		resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
			"timeout": sTimeout,
			"limit":   sLimit,
		}).Get(dtmutil.DefaultHTTPServer + "/resetCronTime")

		m := map[string]interface{}{}
		dtmimp.MustUnmarshalString(resp.String(), &m)
		hasRemaining, ok := m["has_remaining"].(bool)
		assert.Equal(t, ok, true)
		succeedCount, ok := m["succeed_count"].(float64)
		assert.Equal(t, ok, true)
		return int64(succeedCount), hasRemaining, err
	})
}

func TestAPIForceStoppedNormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	busi.MainSwitch.TransOutResult.SetOnce("ONGOING")
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))

	resp, err := dtmcli.GetRestyClient().R().SetBody(map[string]string{
		"gid": saga.Gid,
	}).Post(dtmutil.DefaultHTTPServer + "/forceStop")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode(), http.StatusOK)
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
}

func TestAPIForceStoppedAbnormal(t *testing.T) {
	saga := genSaga(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))

	resp, err := dtmcli.GetRestyClient().R().SetBody(map[string]string{
		"gid": saga.Gid,
	}).Post(dtmutil.DefaultHTTPServer + "/forceStop")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode(), http.StatusConflict)
}

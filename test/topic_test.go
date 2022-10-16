package test

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	testTopicTestTopicNormal           = "test_topic_TestTopicNormal"
	testTopicTestConcurrentUpdateTopic = "concurrent_topic_TestConcurrentUpdateTopic"
)

func TestTopicNormal(t *testing.T) {
	testSubscribe(t, httpSubscribe)
	testUnsubscribe(t, httpUnsubscribe)
	testDeleteTopic(t, httpDeleteTopic)

	testSubscribe(t, grpcSubscribe)
	testUnsubscribe(t, grpcUnsubscribe)
	testDeleteTopic(t, grpcDeleteTopic)
}

func TestConcurrentUpdateTopic(t *testing.T) {
	var wg sync.WaitGroup
	var urls []string
	var errNum int
	concurrentTimes := 20
	// concurrently updates the topic, part of them succeed
	for i := 0; i < concurrentTimes; i++ {
		wg.Add(1)
		go func(i int) {
			url := "http://dtm/test" + strconv.Itoa(i)
			err := httpSubscribe(testTopicTestConcurrentUpdateTopic, url)
			if err == nil {
				urls = append(urls, url)
			} else {
				errNum++
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	assert.True(t, len(urls) > 0)

	// delete successfully subscribed urls above, all of them should succeed
	for _, url := range urls {
		assert.Nil(t, httpUnsubscribe(testTopicTestConcurrentUpdateTopic, url))
	}

	// finally, the topic version should be correct
	m := map[string]interface{}{}
	resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"cat": "topics",
		"key": testTopicTestConcurrentUpdateTopic,
	}).Get(dtmutil.DefaultHTTPServer + "/queryKV")
	assert.Nil(t, err)
	dtmimp.MustUnmarshalString(resp.String(), &m)
	dtmimp.MustRemarshal(m["kv"].([]interface{})[0], &m)
	assert.Equal(t, float64((concurrentTimes-errNum)*2), m["version"])
}

func testSubscribe(t *testing.T, subscribe func(topic, url string) error) {
	assert.Nil(t, subscribe(testTopicTestTopicNormal, "http://dtm/test1"))
	assert.Error(t, subscribe(testTopicTestTopicNormal, "http://dtm/test1")) // error:repeat subscription
	assert.Error(t, subscribe("", "http://dtm/test1"))                       // error:empty topic
	assert.Error(t, subscribe(testTopicTestTopicNormal, ""))                 // error:empty url
	assert.Nil(t, subscribe(testTopicTestTopicNormal, "http://dtm/test2"))
}

func testUnsubscribe(t *testing.T, unsubscribe func(topic, url string) error) {
	assert.Nil(t, unsubscribe(testTopicTestTopicNormal, "http://dtm/test1"))
	assert.Error(t, unsubscribe(testTopicTestTopicNormal, "http://dtm/test1")) // error:repeat unsubscription
	assert.Error(t, unsubscribe("", "http://dtm/test1"))                       // error:empty topic
	assert.Error(t, unsubscribe(testTopicTestTopicNormal, ""))                 // error:empty url
	assert.Error(t, unsubscribe("non_existent_topic", "http://dtm/test1"))     // error:unsubscribe a non-existent topic
	assert.Nil(t, unsubscribe(testTopicTestTopicNormal, "http://dtm/test2"))
	assert.Error(t, unsubscribe(testTopicTestTopicNormal, "http://dtm/test2"))
}

func testDeleteTopic(t *testing.T, deleteTopic func(topic string) error) {
	assert.Error(t, deleteTopic("non_existent_testDeleteTopic"))
	assert.Nil(t, deleteTopic(testTopicTestTopicNormal))
}

func httpSubscribe(topic, url string) error {
	resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"topic":  topic,
		"url":    url,
		"remark": "for test",
	}).Get(dtmutil.DefaultHTTPServer + "/subscribe")
	e2p(err)
	if resp.StatusCode() != 200 {
		err = errors.Errorf("Http Request Error. Resp:%v", resp.String())
	}
	return err
}

func httpUnsubscribe(topic, url string) error {
	resp, err := dtmcli.GetRestyClient().R().SetQueryParams(map[string]string{
		"topic": topic,
		"url":   url,
	}).Get(dtmutil.DefaultHTTPServer + "/unsubscribe")
	e2p(err)
	if resp.StatusCode() != 200 {
		err = errors.Errorf("Http Request Error. Resp:%+v", resp.String())
	}
	return err
}

func httpDeleteTopic(topic string) error {
	resp, err := dtmcli.GetRestyClient().R().Delete(dtmutil.DefaultHTTPServer + "/topic/" + topic)
	e2p(err)
	if resp.StatusCode() != 200 {
		err = errors.Errorf("Http Request Error. Resp:%+v", resp.String())
	}
	return err
}

func grpcSubscribe(topic, url string) error {
	_, err := dtmgimp.MustGetDtmClient(dtmutil.DefaultGrpcServer).Subscribe(context.Background(),
		&dtmgpb.DtmTopicRequest{
			Topic:  topic,
			URL:    url,
			Remark: "for test"})
	return err
}

func grpcUnsubscribe(topic, url string) error {
	_, err := dtmgimp.MustGetDtmClient(dtmutil.DefaultGrpcServer).Unsubscribe(context.Background(),
		&dtmgpb.DtmTopicRequest{
			Topic: topic,
			URL:   url})
	return err
}

func grpcDeleteTopic(topic string) error {
	_, err := dtmgimp.MustGetDtmClient(dtmutil.DefaultGrpcServer).DeleteTopic(context.Background(),
		&dtmgpb.DtmTopicRequest{
			Topic: topic})
	return err
}

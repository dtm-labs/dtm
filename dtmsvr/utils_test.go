/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestUtils(t *testing.T) {
	CronExpiredTrans(1)
	sleepCronTime()
}

func TestSetNextCron(t *testing.T) {
	conf.RetryInterval = 10
	tg := TransGlobal{}
	tg.NextCronInterval = conf.RetryInterval
	tg.RetryInterval = 15
	assert.Equal(t, int64(15), tg.getNextCronInterval(cronReset))
	tg.RetryInterval = 0
	assert.Equal(t, conf.RetryInterval, tg.getNextCronInterval(cronReset))
	assert.Equal(t, conf.RetryInterval*2, tg.getNextCronInterval(cronBackoff))
	tg.TimeoutToFail = 3
	assert.Equal(t, int64(3), tg.getNextCronInterval(cronReset))
}

func TestCopyContext(t *testing.T) {
	var key testContextType = "key"
	var value testContextType = "value"
	ctxWithValue := context.WithValue(context.Background(), key, value)
	newCtx := CopyContext(ctxWithValue)
	assert.Equal(t, ctxWithValue.Value(key), newCtx.Value(key))

	var ctx context.Context
	newCtx = CopyContext(ctx)
	assert.Nil(t, newCtx)
}

type testContextType string

func TestCopyContextRecursive(t *testing.T) {
	var key testContextType = "key"
	var key2 testContextType = "key2"
	var key3 testContextType = "key3"
	var value testContextType = "value"
	var value2 testContextType = "value2"
	var value3 testContextType = "value3"
	var nestedKey testContextType = "nested_key"
	var nestedValue testContextType = "nested_value"
	ctxWithValue := context.WithValue(context.Background(), key, value)
	nestedCtx := context.WithValue(ctxWithValue, nestedKey, nestedValue)
	cancelCtxx, cancel := context.WithCancel(nestedCtx)
	defer cancel()
	timerCtxx, cancel2 := context.WithTimeout(cancelCtxx, time.Duration(10)*time.Second)
	defer cancel2()
	timer2 := context.WithValue(timerCtxx, key2, value2)
	timer3 := context.WithValue(timer2, key3, value3)
	newCtx := CopyContext(timer3)

	assert.Equal(t, timer3.Value(nestedKey), newCtx.Value(nestedKey))
	assert.Equal(t, timer3.Value(key), newCtx.Value(key))
	assert.Equal(t, timer3.Value(key2), newCtx.Value(key2))
	assert.Equal(t, timer3.Value(key3), newCtx.Value(key3))
}

func TestCopyContextWithMetadata(t *testing.T) {
	md := metadata.New(map[string]string{"key": "value"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = metadata.NewOutgoingContext(ctx, md)
	newCtx := CopyContext(ctx)

	copiedMD, ok := metadata.FromIncomingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 1, len(copiedMD["key"]))
	assert.Equal(t, "value", copiedMD["key"][0])
	copiedMD, ok = metadata.FromOutgoingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 1, len(copiedMD["key"]))
	assert.Equal(t, "value", copiedMD["key"][0])
}

func BenchmarkCopyContext(b *testing.B) {
	var key testContextType = "key"
	var value testContextType = "value"
	ctx := context.WithValue(context.Background(), key, value)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CopyContext(ctx)
	}
}

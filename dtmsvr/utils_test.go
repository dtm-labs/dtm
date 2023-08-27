/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"testing"

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
	ctxWithValue := context.WithValue(context.Background(), "key", "value")
	newCtx := CopyContext(ctxWithValue)
	assert.Equal(t, ctxWithValue.Value("key"), newCtx.Value("key"))

	var ctx context.Context
	newCtx = CopyContext(ctx)
	assert.Nil(t, newCtx)
}

func TestCopyContextRecursive(t *testing.T) {
	ctxWithValue := context.WithValue(context.Background(), "key", "value")
	nestedCtx := context.WithValue(ctxWithValue, "nested_key", "nested_value")
	newCtx := CopyContext(nestedCtx)

	assert.Equal(t, nestedCtx.Value("nested_key"), newCtx.Value("nested_key"))
	assert.Equal(t, nestedCtx.Value("key"), newCtx.Value("key"))
}

func TestCopyContextWithMetadata(t *testing.T) {
	md := metadata.New(map[string]string{"key": "value"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx := CopyContext(ctx)

	copiedMD, ok := metadata.FromIncomingContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 1, len(copiedMD["key"]))
	assert.Equal(t, "value", copiedMD["key"][0])
}

func BenchmarkCopyContext(b *testing.B) {
	ctx := context.WithValue(context.Background(), "key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CopyContext(ctx)
	}
}

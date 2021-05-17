package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/dtmsvr"
)

func init() {
	dtmsvr.LoadConfig()
}

func TestViper(t *testing.T) {
	assert.Equal(t, "test_val", viper.GetString("test"))
}

func TTestDtmSvr(t *testing.T) {
	// 发送Prepare请求后，验证数据库
	// ConsumeHalfMsg 验证数据库
	// ConsumeMsg 验证数据库
}

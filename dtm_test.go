package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/spf13/viper"
	"github.com/yedf/dtm/common"
)

func TestCtxKey(t *testing.T) {
	common.LoadConfig()
	assert.Equal(t, "http://localhost:8080/api/dtm/", viper.GetString("tc"))
}

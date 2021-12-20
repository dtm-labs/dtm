package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"testing"
)

func TestDtmMetrics(t *testing.T) {
	rest, err := dtmimp.RestyClient.R().Get("http://localhost:36789/api/metrics")
	assert.Nil(t, err)
	assert.Equal(t, rest.StatusCode(), 200)
}

package dtmsvr

import (
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"testing"
)

func TestDtmMetrics(t *testing.T) {
	common.MustLoadConfig()
	dtmcli.SetCurrentDBType(common.Config.ExamplesDB.Driver)
	PopulateDB(true)
	StartSvr()
	rest, err := dtmimp.RestyClient.R().Get("http://localhost:36789/api/metrics")
	assert.Nil(t, err)
	assert.Equal(t, rest.StatusCode(), 200)
}

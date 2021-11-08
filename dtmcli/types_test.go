package dtmcli

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func TestTypes(t *testing.T) {
	err := dtmimp.CatchP(func() {
		MustGenGid("http://localhost:8080/api/no")
	})
	assert.Error(t, err)
	assert.Error(t, err)
	_, err = BarrierFromQuery(url.Values{})
	assert.Error(t, err)

}

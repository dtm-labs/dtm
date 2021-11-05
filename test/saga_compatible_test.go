package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestSagaCompatibleNormal(t *testing.T) { // compatible with old http, which put payload in steps.data
	gid := dtmimp.GetFuncName()
	body := fmt.Sprintf(`{"gid":"%s","trans_type":"saga","steps":[{"action":"%s/TransOut","compensate":"%s/TransOutRevert","data":"{\"amount\":30,\"transInResult\":\"SUCCESS\",\"transOutResult\":\"SUCCESS\"}"},{"action":"%s/TransIn","compensate":"%s/TransInRevert","data":"{\"amount\":30,\"transInResult\":\"SUCCESS\",\"transOutResult\":\"SUCCESS\"}"}]}`,
		gid, examples.Busi, examples.Busi, examples.Busi, examples.Busi)
	dtmimp.RestyClient.R().SetBody(body).Post(fmt.Sprintf("%s/submit", examples.DtmServer))
	waitTransProcessed(gid)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusSucceed}, getBranchesStatus(gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(gid))
}

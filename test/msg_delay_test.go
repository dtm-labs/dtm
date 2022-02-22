package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func genMsgDelay(gid string) *dtmcli.Msg {
	req := busi.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(dtmutil.DefaultHTTPServer, gid).
		Add(busi.Busi+"/TransOut", &req).
		Add(busi.Busi+"/TransIn", &req).EnableDelay(10)
	msg.QueryPrepared = busi.Busi + "/QueryPrepared"
	return msg
}

func TestMsgDelayNormal(t *testing.T) {
	gid := dtmimp.GetFuncName()
	msg := genMsgDelay(gid)
	msg.Submit()
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnceForwardCron(t, "", 0)
	cronTransOnceForwardCron(t, "", 8)
	cronTransOnceForwardCron(t, gid, 12)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestMsgOptionsTimeout(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	cronTransOnce()
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutCustom(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.TimeoutToFail = 120
	msg.Prepare("")
	cronTransOnce()
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOptionsTimeoutFailed(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.TimeoutToFail = 120
	msg.Prepare("")
	cronTransOnce()
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	cronTransOnceForwardNow(60)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestMsgNormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Submit()
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOngoingSuccess(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgOngoingFailed(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(180)
	assert.Equal(t, []string{dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(msg.Gid))
}

func TestMsgAbnormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	err := msg.Prepare("")
	assert.Nil(t, err)
	err = msg.Submit()
	assert.Nil(t, err)
	err = msg.Submit()
	assert.Nil(t, err)

	err = msg.Prepare("")
	assert.Error(t, err)
}

func genMsg(gid string) *dtmcli.Msg {
	req := examples.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(examples.DtmServer, gid).
		Add(examples.Busi+"/TransOut", &req).
		Add(examples.Busi+"/TransIn", &req)
	msg.QueryPrepared = examples.Busi + "/CanSubmit"
	return msg
}

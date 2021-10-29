package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestMsg(t *testing.T) {

	msgNormal(t)
	msgOngoing(t)
	msgOngoingFailed(t)
}

func msgNormal(t *testing.T) {
	msg := genMsg("gid-msg-normal")
	msg.Submit()
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func msgOngoing(t *testing.T) {
	msg := genMsg("gid-msg-normal-pending")
	msg.Prepare("")
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	CronTransOnce()
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func msgOngoingFailed(t *testing.T) {
	msg := genMsg("gid-msg-pending-failed")
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

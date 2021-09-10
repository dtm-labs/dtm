package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestMsg(t *testing.T) {

	msgNormal(t)
	msgPending(t)
}

func msgNormal(t *testing.T) {
	msg := genMsg("gid-msg-normal")
	msg.Submit()
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	WaitTransProcessed(msg.Gid)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

func msgPending(t *testing.T) {
	msg := genMsg("gid-msg-normal-pending")
	msg.Prepare("")
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, dtmcli.StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce("PENDING")
	CronTransOnce(60 * time.Second)
	assert.Equal(t, dtmcli.StatusSubmitted, getTransStatus(msg.Gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(msg.Gid))
}

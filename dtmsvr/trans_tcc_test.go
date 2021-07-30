package dtmsvr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestTcc(t *testing.T) {
	tccNormal(t)
	tccRollback(t)

}

func tccNormal(t *testing.T) {
	data := &examples.TransReq{Amount: 30}
	gid := "tccNormal"
	err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (rerr error) {
		_, rerr = tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return
	})
	e2p(err)
}

func tccRollback(t *testing.T) {
	gid := "tccRollback"
	data := &examples.TransReq{Amount: 30, TransInResult: "FAILURE"}
	err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (rerr error) {
		resp, rerr := tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Contains(t, resp.String(), "SUCCESS")
		examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		assert.Error(t, rerr)
		return
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, "aborting", getTransStatus(gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "failed", getTransStatus(gid))
}

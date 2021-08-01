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
	ret, err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (interface{}, error) {
		resp, err := tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	dtmcli.PanicIfFailure(ret, err)
}

func tccRollback(t *testing.T) {
	gid := "tccRollback"
	data := &examples.TransReq{Amount: 30, TransInResult: "FAILURE"}
	resp, err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (interface{}, error) {
		resp, rerr := tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.True(t, !dtmcli.IsFailure(resp, rerr))
		examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
		return tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.True(t, dtmcli.IsFailure(resp, err))
	WaitTransProcessed(gid)
	assert.Equal(t, "aborting", getTransStatus(gid))
	CronTransOnce(60 * time.Second)
	assert.Equal(t, "failed", getTransStatus(gid))
}

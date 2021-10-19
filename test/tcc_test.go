package test

import (
	"testing"

	"github.com/go-resty/resty/v2"
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
	err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, err)
		return tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Nil(t, err)
}

func tccRollback(t *testing.T) {
	gid := "tccRollback"
	data := &examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure}
	err := dtmcli.TccGlobalTransaction(examples.DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, rerr := tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		assert.Nil(t, rerr)
		examples.MainSwitch.TransOutRevertResult.SetOnce("PENDING")
		return tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, "aborting", getTransStatus(gid))
	CronTransOnce()
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
}

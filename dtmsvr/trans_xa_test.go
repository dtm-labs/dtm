package dtmsvr

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestXa(t *testing.T) {
	if config.DB["driver"] != "mysql" {
		return
	}
	xaLocalError(t)
	xaNormal(t)
	xaRollback(t)
}

func xaLocalError(t *testing.T) {
	_, err := examples.XaClient.XaGlobalTransaction("xaLocalError", func(xa *dtmcli.Xa) (*resty.Response, error) {
		return nil, fmt.Errorf("an error")
	})
	assert.Error(t, err, fmt.Errorf("an error"))
}

func xaNormal(t *testing.T) {
	xc := examples.XaClient
	gid := "xaNormal"
	res, err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := examples.GenTransReq(30, false, false)
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return xa.CallBranch(req, examples.Busi+"/TransInXa")
	})
	dtmcli.PanicIfFailure(res, err)
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"prepared", "succeed", "prepared", "succeed"}, getBranchesStatus(gid))
}

func xaRollback(t *testing.T) {
	xc := examples.XaClient
	gid := "xaRollback"
	res, err := xc.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := &examples.TransReq{Amount: 30, TransInResult: "FAILURE"}
		resp, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		if dtmcli.IsFailure(resp, err) {
			return resp, err
		}
		return xa.CallBranch(req, examples.Busi+"/TransInXa")
	})
	assert.True(t, dtmcli.IsFailure(res, err))
	WaitTransProcessed(gid)
	assert.Equal(t, []string{"succeed", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}

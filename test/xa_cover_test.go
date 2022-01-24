package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestXaCoverDBError(t *testing.T) {
	oldDriver := getXc().Conf.Driver
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := busi.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		assert.Nil(t, err)
		getXc().Conf.Driver = "no-driver"
		_, err = xa.CallBranch(req, busi.Busi+"/TransInXa")
		assert.Error(t, err)
		return nil, err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	getXc().Conf.Driver = oldDriver
	cronTransOnceForwardNow(t, gid, 500) // rollback succeeded here
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

func TestXaCoverDTMError(t *testing.T) {
	oldServer := getXc().Server
	getXc().Server = "localhost:01"
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		return nil, nil
	})
	assert.Error(t, err)
	getXc().Server = oldServer
}

func TestXaCoverGidError(t *testing.T) {
	gid := dtmimp.GetFuncName() + "-'  '"
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := busi.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, busi.Busi+"/TransOutXa")
		assert.Error(t, err)
		return nil, err
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
}

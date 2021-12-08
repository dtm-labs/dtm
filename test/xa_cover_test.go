package test

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestXaCoverDBError(t *testing.T) {
	oldDriver := getXc().Conf["driver"]
	gid := dtmimp.GetFuncName()
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := examples.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		assert.Nil(t, err)
		getXc().Conf["driver"] = "no-driver"
		_, err = xa.CallBranch(req, examples.Busi+"/TransInXa")
		assert.Error(t, err)
		getXc().Conf["driver"] = oldDriver // make abort succeed
		return nil, err
	})
	assert.Error(t, err)
	getXc().Conf["driver"] = "no-driver" // make xa rollback failed
	waitTransProcessed(gid)
	getXc().Conf["driver"] = oldDriver
	cronTransOnceForwardNow(500) // rollback succeeded here
	assert.Equal(t, StatusFailed, getTransStatus(gid))
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
	gid := "errgid-'  '"
	err := getXc().XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
		req := examples.GenTransReq(30, false, false)
		_, err := xa.CallBranch(req, examples.Busi+"/TransOutXa")
		assert.Error(t, err)
		return nil, err
	})
	assert.Error(t, err)
}

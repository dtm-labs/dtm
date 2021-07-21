package dtmsvr

import (
	"testing"

	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestTcc(t *testing.T) {
	tccNormal(t)
	tccRollback(t)

}

func tccNormal(t *testing.T) {
	data := &examples.TransReq{Amount: 30}
	_, err := dtmcli.TccGlobalTransaction(examples.DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		_, rerr = tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return
	})
	e2p(err)
}

func tccRollback(t *testing.T) {
	data := &examples.TransReq{Amount: 30, TransInResult: "FAILURE"}
	_, err := dtmcli.TccGlobalTransaction(examples.DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		_, rerr = tcc.CallBranch(data, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
		e2p(rerr)
		_, rerr = tcc.CallBranch(data, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		e2p(rerr)
		return
	})
	e2p(err)
}

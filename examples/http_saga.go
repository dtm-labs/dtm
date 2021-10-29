package examples

import (
	"github.com/yedf/dtm/dtmcli"
)

func init() {
	addSample("saga", func() string {
		dtmcli.Logf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		dtmcli.Logf("saga busi trans submit")
		err := saga.Submit()
		dtmcli.Logf("result gid is: %s", saga.Gid)
		dtmcli.FatalIfError(err)
		return saga.Gid
	})
	addSample("saga_wait", func() string {
		dtmcli.Logf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
		err := saga.Submit()
		dtmcli.Logf("result gid is: %s", saga.Gid)
		dtmcli.FatalIfError(err)
		return saga.Gid
	})

}

package examples

import (
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func init() {
	addSample("saga", func() string {
		dtmimp.Logf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		dtmimp.Logf("saga busi trans submit")
		err := saga.Submit()
		dtmimp.Logf("result gid is: %s", saga.Gid)
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
	addSample("saga_wait", func() string {
		dtmimp.Logf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		saga.SetOptions(&dtmimp.TransOptions{WaitResult: true})
		err := saga.Submit()
		dtmimp.Logf("result gid is: %s", saga.Gid)
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
	addSample("concurrent_saga", func() string {
		dtmimp.Logf("a concurrent saga busi transaction begin")
		req := &TransReq{Amount: 30}
		csaga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req).
			EnableConcurrent().
			AddBranchOrder(2, []int{0, 1}).
			AddBranchOrder(3, []int{0, 1})
		dtmimp.Logf("concurrent saga busi trans submit")
		err := csaga.Submit()
		dtmimp.Logf("result gid is: %s", csaga.Gid)
		dtmimp.FatalIfError(err)
		return csaga.Gid
	})
}

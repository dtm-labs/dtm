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
		saga.WaitResult = true // 设置为等待结果模式，后面的submit调用，会等待服务器处理这个事务。如果Submit正常返回，那么整个全局事务已成功完成
		err := saga.Submit()
		dtmcli.Logf("result gid is: %s", saga.Gid)
		dtmcli.FatalIfError(err)
		return saga.Gid
	})

}

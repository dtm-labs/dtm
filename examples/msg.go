package examples

import (
	"github.com/yedf/dtm/dtmcli"
)

// MsgFireRequest 1
func MsgFireRequest() string {
	dtmcli.Logf("a busi transaction begin")
	req := &TransReq{Amount: 30}
	msg := dtmcli.NewMsg(DtmServer, dtmcli.MustGenGid(DtmServer)).
		Add(Busi+"/TransOut", req).
		Add(Busi+"/TransIn", req)
	err := msg.Prepare(Busi + "/TransQuery")
	dtmcli.FatalIfError(err)
	dtmcli.Logf("busi trans submit")
	err = msg.Submit()
	dtmcli.FatalIfError(err)
	return msg.Gid
}

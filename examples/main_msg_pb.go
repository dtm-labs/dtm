package examples

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
	dtmpb "github.com/yedf/dtm/dtmpb"
)

// MsgPbSetup 1
func MsgPbSetup(app *gin.Engine) {

}

// MsgPbFireRequest 1
func MsgPbFireRequest() string {
	dtmcli.Logf("MsgPbFireRequest")
	reply, err := DtmClient.Call(context.Background(), &dtmpb.DtmRequest{
		Gid:       "pb_test",
		TransType: "msg",
		Method:    "submit",
		Extra: dtmcli.MS{
			"BusiFunc": BusiPb + "/examples.Busi/Call",
		},
	})
	dtmcli.Logf("reply and err is: %v, %v", reply, err)
	return ""
}

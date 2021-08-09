package examples

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/dtmcli"
)

// MsgPbSetup 1
func MsgPbSetup(app *gin.Engine) {

}

// MsgPbFireRequest 1
func MsgPbFireRequest() string {
	dtmcli.Logf("MsgPbFireRequest")
	reply, err := DtmClient.Call(context.Background(), &dtmcli.DtmRequest{})
	dtmcli.Logf("reply and err is: %v, %v", reply, err)
	return ""
}

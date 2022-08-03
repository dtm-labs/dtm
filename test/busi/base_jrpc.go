package busi

import (
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
)

// BusiJrpcURL url prefix for busi
var BusiJrpcURL = fmt.Sprintf("http://localhost:%d/api/json-rpc?method=", BusiPort)

func addJrpcRoute(app *gin.Engine) {
	app.POST("/api/json-rpc", dtmutil.WrapHandler(func(c *gin.Context) interface{} {
		var data map[string]interface{}
		err := c.BindJSON(&data)
		dtmimp.E2P(err)
		logger.Debugf("method is: %s", data["method"])
		var rerr map[string]interface{}
		r := MainSwitch.JrpcResult.Fetch()
		if r != "" {
			rerr = map[string]interface{}{
				"code": map[string]int{
					"FAILURE": dtmimp.JrpcCodeFailure,
					"ONGOING": dtmimp.JrpcCodeOngoing,
					"OTHER":   -23977,
				},
			}
		}
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"error":   rerr,
			"id":      data["id"],
		}
	}))
}

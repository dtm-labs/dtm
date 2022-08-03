package dtmsvr

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
)

type jrpcReq struct {
	Method  string      `json:"method"`
	Jsonrpc string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

func addJrpcRouter(engine *gin.Engine) {
	type jrpcFunc = func(interface{}) interface{}
	handlers := map[string]jrpcFunc{
		"newGid":         jrpcNewGid,
		"prepare":        jrpcPrepare,
		"submit":         jrpcSubmit,
		"abort":          jrpcAbort,
		"registerBranch": jrpcRegisterBranch,
	}
	engine.POST("/api/json-rpc", func(c *gin.Context) {
		began := time.Now()
		var err error
		var req jrpcReq
		var jerr map[string]interface{}
		r := func() interface{} {
			defer dtmimp.P2E(&err)
			err2 := c.BindJSON(&req)
			if err2 != nil {
				jerr = map[string]interface{}{
					"code":    -32700,
					"message": fmt.Sprintf("Parse json error: %s", err2.Error()),
				}
			} else if req.ID == "" || req.Jsonrpc != "2.0" {
				jerr = map[string]interface{}{
					"code":    -32600,
					"message": fmt.Sprintf("Bad json request: %s", dtmimp.MustMarshalString(req)),
				}
			} else if handlers[req.Method] == nil {
				jerr = map[string]interface{}{
					"code":    -32601,
					"message": fmt.Sprintf("Method not found: %s", req.Method),
				}
			} else if handlers[req.Method] != nil {
				return handlers[req.Method](req.Params)
			}
			return nil
		}()

		// error maybe returned in r, assign it to err
		if ne, ok := r.(error); ok && err == nil {
			err = ne
		}

		if err != nil {
			if errors.Is(err, dtmcli.ErrFailure) {
				jerr = map[string]interface{}{
					"code":    dtmimp.JrpcCodeFailure,
					"message": err.Error(),
				}
				//// following is commented for server
				// } else if errors.Is(err, dtmcli.ErrOngoing) {
				// 	jerr = map[string]interface{}{
				// 		"code":    jrpcCodeOngoing,
				// 		"message": err.Error(),
				// 	}
			} else if jerr == nil {
				jerr = map[string]interface{}{
					"code":    -32603,
					"message": err.Error(),
				}
			}
		}

		result := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"error":   jerr,
			"result":  r,
		}
		b, _ := json.Marshal(result)
		cont := string(b)
		if jerr == nil || jerr["code"] == dtmimp.JrpcCodeOngoing {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), 200, c.Request.Method, c.Request.RequestURI, cont)
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), 200, c.Request.Method, c.Request.RequestURI, cont)
		}
		c.JSON(200, result)
	})
}

// TransFromJrpcParams construct TransGlobal from jrpc params
func TransFromJrpcParams(params interface{}) *TransGlobal {
	t := TransGlobal{}
	dtmimp.MustRemarshal(params, &t)
	t.setupPayloads()
	return &t
}

func jrpcNewGid(interface{}) interface{} {
	return map[string]interface{}{"gid": GenGid()}
}

func jrpcPrepare(params interface{}) interface{} {
	return svcPrepare(TransFromJrpcParams(params))
}

func jrpcSubmit(params interface{}) interface{} {
	return svcSubmit(TransFromJrpcParams(params))
}

func jrpcAbort(params interface{}) interface{} {
	return svcAbort(TransFromJrpcParams(params))
}

func jrpcRegisterBranch(params interface{}) interface{} {
	data := map[string]string{}
	dtmimp.MustRemarshal(params, &data)
	branch := TransBranch{
		Gid:      data["gid"],
		BranchID: data["branch_id"],
		Status:   dtmcli.StatusPrepared,
		BinData:  []byte(data["data"]),
	}
	return svcRegisterBranch(data["trans_type"], &branch, data)
}

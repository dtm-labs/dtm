package dtmsvr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

type jsonRPCReq struct {
	Method  string      `json:"method"`
	Jsonrpc string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

func addJSONRPCRouter(engine *gin.Engine) {
	engine.POST("/", dispatcher)
}

func dispatcher(c *gin.Context) {
	req := new(jsonRPCReq)
	err := c.BindJSON(req)
	logger.Infof("request:%s\n", req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": nil, "error": map[string]interface{}{"code": -32700, "message": "Parse error"}})
		return
	}
	if req.Method == "dtmserver.NewGid" {
		res := jsonRPCNewGid()
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": res, "error": err})
		return
	}

	if req.Method == "dtmserver.Prepare" {
		res := jsonRPCPrepare(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.Submit" {
		res := jsonRPCSubmit(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.Abort" {
		res := jsonRPCAbort(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.RegisterBranch" {
		res := jsonRPCRegisterBranch(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": res, "error": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": req.ID, "result": nil, "error": map[string]interface{}{"code": -32601, "message": "Method not found"}})
}

func jsonRPCNewGid() interface{} {
	return map[string]interface{}{"gid": GenGid(), "dtm_result": dtmcli.ResultSuccess}
}

func jsonRPCPrepare(params interface{}) interface{} {
	res := svcPrepare(TransFromJSONRPCContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRPCSubmit(params interface{}) interface{} {
	res := svcSubmit(TransFromJSONRPCContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRPCAbort(params interface{}) interface{} {
	res := svcAbort(TransFromJSONRPCContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRPCRegisterBranch(params interface{}) interface{} {
	data := map[string]string{}
	paramsJSON, _ := json.Marshal(params)
	err := json.Unmarshal(paramsJSON, &data)
	if err != nil {
		return map[string]string{"dtm_result": "FAILURE", "message": err.Error()}
	}
	branch := TransBranch{
		Gid:      data["gid"],
		BranchID: data["branch_id"],
		Status:   dtmcli.StatusPrepared,
		BinData:  []byte(data["data"]),
	}
	res := svcRegisterBranch(data["trans_type"], &branch, data)
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": res.Error()}
}

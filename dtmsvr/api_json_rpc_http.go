package dtmsvr

import (
	"encoding/json"
	"fmt"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type jsonRpcHttpReq struct {
	Method  string      `json:"method"`
	Jsonrpc string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
	Id      string      `json:"id"`
}

func addJsonRpcHttpRouter(engine *gin.Engine) {
	engine.POST("/", dispatcher)
}

func dispatcher(c *gin.Context) {
	req := new(jsonRpcHttpReq)
	err := c.BindJSON(req)
	logger.Infof("request:%s\n", req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": nil, "error": map[string]interface{}{"code": -32700, "message": "Parse error"}})
		return
	}
	if req.Method == "dtmserver.NewGid" {
		res, err := jsonRpcHttpNewGid()
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": res, "error": err})
		return
	}

	if req.Method == "dtmserver.Prepare" {
		res := jsonRpcHttpPrepare(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.Submit" {
		res := jsonRpcHttpSubmit(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.Abort" {
		res := jsonRpcHttpAbort(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": res, "error": nil})
		return
	}

	if req.Method == "dtmserver.RegisterBranch" {
		res := jsonRpcHttpRegisterBranch(req.Params)
		c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": res, "error": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": req.Id, "result": nil, "error": map[string]interface{}{"code": -32601, "message": "Method not found"}})
	return
}

func jsonRpcHttpNewGid() (interface{}, error) {
	return map[string]interface{}{"gid": GenGid(), "dtm_result": dtmcli.ResultSuccess}, nil
}

func jsonRpcHttpPrepare(params interface{}) interface{} {
	res := svcPrepare(TransFromJsonRpcHttpContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRpcHttpSubmit(params interface{}) interface{} {
	res := svcSubmit(TransFromJsonRpcHttpContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRpcHttpAbort(params interface{}) interface{} {
	res := svcAbort(TransFromJsonRpcHttpContext(params))
	if res == nil {
		return map[string]string{"dtm_result": "SUCCESS"}
	}
	return map[string]string{"dtm_result": "FAILURE", "message": fmt.Sprintf("%v", res)}
}

func jsonRpcHttpRegisterBranch(params interface{}) interface{} {
	data := map[string]string{}
	paramsJson, _ := json.Marshal(params)
	err := json.Unmarshal(paramsJson, &data)
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

package examples

import (
	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
)

var e2p = common.E2P

type M = map[string]interface{}

// 指定dtm服务地址
const DtmServer = "http://localhost:8080/api/dtmsvr"

type TransReq struct {
	Amount         int    `json:"amount"`
	TransInResult  string `json:"transInResult"`
	TransOutResult string `json:"transOutResult"`
}

func GenTransReq(amount int, outFailed bool, inFailed bool) *TransReq {
	return &TransReq{
		Amount:         amount,
		TransOutResult: common.If(outFailed, "FAIL", "SUCCESS").(string),
		TransInResult:  common.If(inFailed, "FAIL", "SUCCESS").(string),
	}
}

func transReqFromContext(c *gin.Context) *TransReq {
	req := TransReq{}
	err := c.BindJSON(&req)
	e2p(err)
	return &req
}

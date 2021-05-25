package examples

import "github.com/yedf/dtm/common"

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

package examples

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

var e2p = common.E2P

// M alias
type M = map[string]interface{}

// DtmServer dtm service address
const DtmServer = "http://localhost:8080/api/dtmsvr"

const (
	// SagaBarrierBusiPort saga barrier sample port
	SagaBarrierBusiPort = iota + 8090
	// TccBarrierBusiPort tcc barrier sample port
	TccBarrierBusiPort
)

// TransReq transaction request payload
type TransReq struct {
	Amount         int    `json:"amount"`
	TransInResult  string `json:"transInResult"`
	TransOutResult string `json:"transOutResult"`
}

func (t *TransReq) String() string {
	return fmt.Sprintf("amount: %d transIn: %s transOut: %s", t.Amount, t.TransInResult, t.TransOutResult)
}

// GenTransReq 1
func GenTransReq(amount int, outFailed bool, inFailed bool) *TransReq {
	return &TransReq{
		Amount:         amount,
		TransOutResult: common.If(outFailed, "FAIL", "SUCCESS").(string),
		TransInResult:  common.If(inFailed, "FAIL", "SUCCESS").(string),
	}
}

func reqFrom(c *gin.Context) *TransReq {
	req := TransReq{}
	err := c.BindJSON(&req)
	e2p(err)
	return &req
}

func infoFromContext(c *gin.Context) *dtmcli.TransInfo {
	info := dtmcli.TransInfo{
		TransType:  c.Query("trans_type"),
		Gid:        c.Query("gid"),
		BranchID:   c.Query("branch_id"),
		BranchType: c.Query("branch_type"),
	}
	return &info
}

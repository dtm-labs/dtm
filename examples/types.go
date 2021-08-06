package examples

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

var e2p = dtmcli.E2P

// M alias
type M = map[string]interface{}

// DtmServer dtm service address
const DtmServer = "http://localhost:8080/api/dtmsvr"

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
		TransOutResult: dtmcli.If(outFailed, "FAILURE", "SUCCESS").(string),
		TransInResult:  dtmcli.If(inFailed, "FAILURE", "SUCCESS").(string),
	}
}

func reqFrom(c *gin.Context) *TransReq {
	v, ok := c.Get("trans_req")
	if !ok {
		req := TransReq{}
		err := c.BindJSON(&req)
		e2p(err)
		c.Set("trans_req", &req)
		v = &req
	}
	return v.(*TransReq)
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

func dbGet() *common.DB {
	var config = common.GetDBConfig()
	return common.DbGet(config.DB)
}

func sdbGet() *sql.DB {
	var config = common.GetDBConfig()
	db, err := dtmcli.SdbGet(config.DB)
	e2p(err)
	return db
}

// MustGetTrans construct transaction info from request
func MustGetTrans(c *gin.Context) *dtmcli.TransInfo {
	ti, err := dtmcli.TransInfoFromQuery(c.Request.URL.Query())
	e2p(err)
	return ti
}

package dtmcli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := common.MS{}
	resp, err := common.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// CheckResponse 检查Response，返回错误
func CheckResponse(resp *resty.Response, err error) error {
	if err == nil && resp != nil {
		if resp.IsError() {
			return errors.New(resp.String())
		} else if strings.Contains(resp.String(), "FAILURE") {
			return ErrFailure
		}
	}
	return err
}

// CheckResult 检查Result，返回错误
func CheckResult(res interface{}, err error) error {
	resp, ok := res.(*resty.Response)
	if ok {
		return CheckResponse(resp, err)
	}
	if res != nil && strings.Contains(common.MustMarshalString(res), "FAILURE") {
		return ErrFailure
	}
	return err
}

// IDGenerator used to generate a branch id
type IDGenerator struct {
	parentID string
	branchID int
}

// NewBranchID generate a branch id
func (g *IDGenerator) NewBranchID() string {
	if g.branchID >= 99 {
		panic(fmt.Errorf("branch id is larger than 99"))
	}
	if len(g.parentID) >= 20 {
		panic(fmt.Errorf("total branch id is longer than 20"))
	}
	g.branchID = g.branchID + 1
	return g.parentID + fmt.Sprintf("%02d", g.branchID)
}

// TransResult dtm 返回的结果
type TransResult struct {
	DtmResult string `json:"dtm_result"`
	Message   string
}

// TransBase 事务的基础类
type TransBase struct {
	IDGenerator
	Dtm string
	// WaitResult 是否等待全局事务的最终结果
	WaitResult bool
}

// TransBaseFromReq construct xa info from request
func TransBaseFromReq(c *gin.Context) *TransBase {
	return &TransBase{
		IDGenerator: IDGenerator{parentID: c.Query("branch_id")},
		Dtm:         c.Query("dtm"),
	}
}

// CallDtm 调用dtm服务器，返回事务的状态
func (tb *TransBase) CallDtm(body interface{}, operation string) error {
	params := common.MS{}
	if tb.WaitResult {
		params["wait_result"] = "1"
	}
	resp, err := common.RestyClient.R().SetQueryParams(params).
		SetResult(&TransResult{}).SetBody(body).Post(fmt.Sprintf("%s/%s", tb.Dtm, operation))
	if err != nil {
		return err
	}
	tr := resp.Result().(*TransResult)
	if tr.DtmResult == "FAILURE" {
		return errors.New("FAILURE: " + tr.Message)
	}
	return nil
}

// ErrFailure 表示返回失败，要求回滚
var ErrFailure = errors.New("transaction FAILURE")

// ResultSuccess 表示返回成功，可以进行下一步
var ResultSuccess = common.M{"dtm_result": "SUCCESS"}

// ResultFailure 表示返回失败，要求回滚
var ResultFailure = common.M{"dtm_result": "FAILURE"}

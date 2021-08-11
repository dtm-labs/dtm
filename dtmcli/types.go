package dtmcli

import (
	"errors"
	"fmt"
	"net/url"
)

// M a short name
type M = map[string]interface{}

// MS a short name
type MS = map[string]string

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := MS{}
	resp, err := RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
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
	Gid       string `json:"gid"`
	TransType string `json:"trans_type"`
	IDGenerator
	Dtm string
	// WaitResult 是否等待全局事务的最终结果
	WaitResult bool
}

// NewTransBase 1
func NewTransBase(gid string, transType string, dtm string, parentID string) *TransBase {
	return &TransBase{
		Gid:         gid,
		TransType:   transType,
		IDGenerator: IDGenerator{parentID: parentID},
		Dtm:         dtm,
	}
}

// TransBaseFromQuery construct transaction info from request
func TransBaseFromQuery(qs url.Values) *TransBase {
	return NewTransBase(qs.Get("gid"), qs.Get("trans_type"), qs.Get("dtm"), qs.Get("branch_id"))
}

// callDtm 调用dtm服务器，返回事务的状态
func (tb *TransBase) callDtm(body interface{}, operation string) error {
	params := MS{}
	if tb.WaitResult {
		params["wait_result"] = "1"
	}
	resp, err := RestyClient.R().SetQueryParams(params).
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

// ErrPending 表示暂时失败，要求重试
var ErrPending = errors.New("transaction PENDING")

// ResultSuccess 表示返回成功，可以进行下一步
var ResultSuccess = M{"dtm_result": "SUCCESS"}

// ResultFailure 表示返回失败，要求回滚
var ResultFailure = M{"dtm_result": "FAILURE"}

package dtmcli

import (
	"errors"
	"fmt"
	"strings"

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

// IsFailure 如果err非空，或者ret是http的响应且包含FAILURE，那么返回true。此时认为业务调用失败
func IsFailure(res interface{}, err error) bool {
	resp, ok := res.(*resty.Response)
	return err != nil || // 包含错误
		ok && (resp.IsError() || strings.Contains(resp.String(), "FAILURE")) || // resp包含failure
		!ok && res != nil && strings.Contains(common.MustMarshalString(res), "FAILURE") // 结果中包含failure
}

// PanicIfFailure 如果err非空，或者ret是http的响应且包含FAILURE，那么Panic。此时认为业务调用失败
func PanicIfFailure(res interface{}, err error) {
	if IsFailure(res, err) {
		panic(fmt.Errorf("dtm failure ret: %v err %v", res, err))
	}
}

// CheckUserResponse 检查Response，返回错误
func CheckUserResponse(resp *resty.Response, err error) error {
	if err == nil && resp != nil {
		if resp.IsError() {
			return errors.New(resp.String())
		} else if strings.Contains(resp.String(), "FAILURE") {
			return ErrUserFailure
		}
	}
	return err
}

// CheckDtmResponse check the response of dtm, if not ok ,generate error
func CheckDtmResponse(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}
	if !strings.Contains(resp.String(), "SUCCESS") || resp.IsError() {
		return fmt.Errorf("dtm response failed: %s", resp.String())
	}
	return nil
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

// TransStatus 全局事务状态，采用string
type TransStatus string

const (
	// TransEmpty 空值
	TransEmpty TransStatus = ""
	// TransSubmitted 已提交给DTM
	TransSubmitted TransStatus = "submitted"
	// TransAborting 正在回滚中，有两种情况会出现，一是用户侧发起abort请求，而是发起submit同步请求，但是dtm进行回滚中出现错误
	TransAborting TransStatus = "aborting"
	// TransSucceed 事务已完成
	TransSucceed TransStatus = "succeed"
	// TransFailed 事务已回滚
	TransFailed TransStatus = "failed"
	// TransErrorPrepare prepare调用报错
	TransErrorPrepare TransStatus = "error_parepare"
	// TransErrorSubmit submit调用报错
	TransErrorSubmit TransStatus = "error_submit"
	// TransErrorAbort abort调用报错
	TransErrorAbort TransStatus = "error_abort"
)

// TransOptions 提交/终止事务的选项
type TransOptions struct {
	// WaitResult 是否等待全局事务的最终结果
	WaitResult bool
}

// TransResult dtm 返回的结果
type TransResult struct {
	DtmResult string `json:"dtm_result"`
	Status    TransStatus
	Message   string
}

func callDtm(dtm string, body interface{}, operation string, opt *TransOptions) (TransStatus, error) {
	resp, err := common.RestyClient.R().SetQueryParams(common.MS{
		"wait_result": common.If(opt.WaitResult, "1", "").(string),
	}).SetResult(&TransResult{}).SetBody(body).Post(fmt.Sprintf("%s/%s", dtm, operation))
	errStatus := TransStatus("error_" + operation)
	if err != nil {
		return errStatus, err
	}
	tr := resp.Result().(*TransResult)
	if tr.DtmResult == "FAILURE" {
		return errStatus, errors.New(tr.Message)
	}
	return tr.Status, nil
}

func callDtmSimple(dtm string, body interface{}, operation string) error {
	_, err := callDtm(dtm, body, operation, &TransOptions{})
	return err
}

// ErrUserFailure 表示用户返回失败，要求回滚
var ErrUserFailure = errors.New("user return FAILURE")

// ErrDtmFailure 表示用户返回失败，要求回滚
var ErrDtmFailure = errors.New("dtm return FAILURE")

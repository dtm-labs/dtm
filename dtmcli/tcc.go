package dtmcli

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

// Tcc struct of tcc
type Tcc struct {
	TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *Tcc) (*resty.Response, error)

// TccGlobalTransaction begin a tcc global transaction
// dtm: dtm server addresss
// gid: global transaction id
// tccFunc: tcc transacion function, it'll define branch transaction
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	tcc := &Tcc{TransBase: *NewTransBase(gid, "tcc", dtm, "")}
	rerr = tcc.callDtm(tcc, "prepare")
	if rerr != nil {
		return rerr
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := If(x == nil && rerr == nil, "submit", "abort").(string)
		err := tcc.callDtm(tcc, operation)
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	resp, rerr := tccFunc(tcc)
	rerr = CheckResponse(resp, rerr)
	return
}

// TccFromQuery tcc from request info
func TccFromQuery(qs url.Values) (*Tcc, error) {
	tcc := &Tcc{TransBase: *TransBaseFromQuery(qs)}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s parentID: %s", tcc.Dtm, tcc.Gid, tcc.parentID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
// 函数首先注册子事务的所有分支，成功后调用try分支，返回try分支的调用结果
func (t *Tcc) CallBranch(body interface{}, tryURL string, confirmURL string, cancelURL string) (*resty.Response, error) {
	branchID := t.NewBranchID()
	err := t.callDtm(&M{
		"gid":         t.Gid,
		"branch_id":   branchID,
		"trans_type":  "tcc",
		"data":        string(MustMarshal(body)),
		BranchTry:     tryURL,
		BranchConfirm: confirmURL,
		"cancel":      cancelURL,
	}, "registerTccBranch")
	if err != nil {
		return nil, err
	}
	resp, err := RestyClient.R().
		SetBody(body).
		SetQueryParams(MS{
			"dtm":         t.Dtm,
			"gid":         t.Gid,
			"branch_id":   branchID,
			"trans_type":  "tcc",
			"branch_type": BranchTry,
		}).
		Post(tryURL)
	return resp, CheckResponse(resp, err)
}

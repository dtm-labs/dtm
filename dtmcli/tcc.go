package dtmcli

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

// Tcc struct of tcc
type Tcc struct {
	dtmimp.TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *Tcc) (*resty.Response, error)

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm服务器地址
// gid 全局事务id
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	tcc := &Tcc{TransBase: *dtmimp.NewTransBase(gid, "tcc", dtm, "")}
	rerr = dtmimp.TransCallDtm(&tcc.TransBase, tcc, "prepare")
	if rerr != nil {
		return rerr
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := dtmimp.If(x == nil && rerr == nil, "submit", "abort").(string)
		err := dtmimp.TransCallDtm(&tcc.TransBase, tcc, operation)
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = tccFunc(tcc)
	return
}

// TccFromQuery tcc from request info
func TccFromQuery(qs url.Values) (*Tcc, error) {
	tcc := &Tcc{TransBase: *dtmimp.TransBaseFromQuery(qs)}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s parentID: %s", tcc.Dtm, tcc.Gid, tcc.BranchID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
func (t *Tcc) CallBranch(body interface{}, tryURL string, confirmURL string, cancelURL string) (*resty.Response, error) {
	branchID := t.NewSubBranchID()
	err := dtmimp.TransRegisterBranch(&t.TransBase, map[string]string{
		"data":        dtmimp.MustMarshalString(body),
		"branch_id":   branchID,
		BranchConfirm: confirmURL,
		BranchCancel:  cancelURL,
	}, "registerTccBranch")
	if err != nil {
		return nil, err
	}
	return dtmimp.TransRequestBranch(&t.TransBase, body, branchID, BranchTry, tryURL)
}

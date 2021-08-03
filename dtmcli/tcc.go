package dtmcli

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

// Tcc struct of tcc
type Tcc struct {
	IDGenerator
	Dtm string
	Gid string
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *Tcc) (*resty.Response, error)

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm服务器地址
// gid 全局事务id
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (ret interface{}, rerr error) {
	data := &M{
		"gid":        gid,
		"trans_type": "tcc",
	}
	tcc := &Tcc{Dtm: dtm, Gid: gid}
	resp, err := CallDtm(dtm, data, "prepare", &TransOptions{})
	if err != nil {
		return resp, err
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := common.If(x == nil && rerr == nil, "submit", "abort").(string)
		resp, err = CallDtm(dtm, data, operation, &TransOptions{})
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	ret, rerr = tccFunc(tcc)
	rerr = CheckResult(ret, rerr)
	return
}

// TccFromReq tcc from request info
func TccFromReq(c *gin.Context) (*Tcc, error) {
	tcc := &Tcc{
		Dtm:         c.Query("dtm"),
		Gid:         c.Query("gid"),
		IDGenerator: IDGenerator{parentID: c.Query("branch_id")},
	}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s", tcc.Dtm, tcc.Gid)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
// 函数首先注册子事务的所有分支，成功后调用try分支，返回try分支的调用结果
func (t *Tcc) CallBranch(body interface{}, tryURL string, confirmURL string, cancelURL string) (*resty.Response, error) {
	branchID := t.NewBranchID()
	_, err := CallDtm(t.Dtm, &M{
		"gid":        t.Gid,
		"branch_id":  branchID,
		"trans_type": "tcc",
		"status":     "prepared",
		"data":       string(common.MustMarshal(body)),
		"try":        tryURL,
		"confirm":    confirmURL,
		"cancel":     cancelURL,
	}, "registerTccBranch", &TransOptions{})
	if err != nil {
		return nil, err
	}
	resp, err := common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"dtm":         t.Dtm,
			"gid":         t.Gid,
			"branch_id":   branchID,
			"trans_type":  "tcc",
			"branch_type": "try",
		}).
		Post(tryURL)
	return resp, CheckResponse(resp, err)
}

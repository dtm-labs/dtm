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
type TccGlobalFunc func(tcc *Tcc) (interface{}, error)

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
	resp, err := common.RestyClient.R().SetBody(data).Post(tcc.Dtm + "/prepare")
	if IsFailure(resp, err) {
		return resp, err
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		var x interface{}
		if x = recover(); x != nil || IsFailure(ret, rerr) {
			resp, err = common.RestyClient.R().SetBody(data).Post(dtm + "/abort")
		} else {
			resp, err = common.RestyClient.R().SetBody(data).Post(dtm + "/submit")
		}
		if IsFailure(resp, err) {
			common.RedLogf("submitting or abort global transaction error: %v resp: %s", err, resp.String())
		}
		if x != nil {
			panic(x)
		}
	}()
	ret, rerr = tccFunc(tcc)
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
	resp, err := common.RestyClient.R().
		SetBody(&M{
			"gid":        t.Gid,
			"branch_id":  branchID,
			"trans_type": "tcc",
			"status":     "prepared",
			"data":       string(common.MustMarshal(body)),
			"try":        tryURL,
			"confirm":    confirmURL,
			"cancel":     cancelURL,
		}).
		Post(t.Dtm + "/registerTccBranch")
	if IsFailure(resp, err) {
		return resp, err
	}
	return common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"dtm":         t.Dtm,
			"gid":         t.Gid,
			"branch_id":   branchID,
			"trans_type":  "tcc",
			"branch_type": "try",
		}).
		Post(tryURL)
}

package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

// M alias
type M = map[string]interface{}

var e2p = common.E2P

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *Xa) (*resty.Response, error)

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *Xa) (interface{}, error)

// XaRegisterCallback type of xa register callback handler
type XaRegisterCallback func(path string, xa *XaClient)

// XaClient xa client
type XaClient struct {
	Server      string
	Conf        map[string]string
	CallbackURL string
}

// Xa xa transaction
type Xa struct {
	IDGenerator
	Gid string
}

// XaFromReq construct xa info from request
func XaFromReq(c *gin.Context) *Xa {
	return &Xa{
		Gid:         c.Query("gid"),
		IDGenerator: IDGenerator{parentID: c.Query("branch_id")},
	}
}

// NewXaClient construct a xa client
func NewXaClient(server string, mysqlConf map[string]string, callbackURL string, register XaRegisterCallback) (*XaClient, error) {
	xa := &XaClient{
		Server:      server,
		Conf:        mysqlConf,
		CallbackURL: callbackURL,
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return nil, err
	}
	register(u.Path, xa)
	return xa, nil
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaClient) HandleCallback(gid string, branchID string, action string) (interface{}, error) {
	db := common.SdbAlone(xc.Conf)
	defer db.Close()
	xaID := gid + "-" + branchID
	_, err := common.SdbExec(db, fmt.Sprintf("xa %s '%s'", action, xaID))
	return ResultSuccess, err

}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(c *gin.Context, xaFunc XaLocalFunc) (ret interface{}, rerr error) {
	xa := XaFromReq(c)
	branchID := xa.NewBranchID()
	xaBranch := xa.Gid + "-" + branchID
	db := common.SdbAlone(xc.Conf)
	defer func() { db.Close() }()
	defer func() {
		x := recover()
		_, err := common.SdbExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
		if x == nil && rerr == nil && err == nil {
			_, err = common.SdbExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
		}
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = common.SdbExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	if rerr != nil {
		return
	}
	ret, rerr = xaFunc(db, xa)
	rerr = CheckResult(ret, rerr)
	if rerr != nil {
		return
	}
	rerr = CallDtm(xc.Server, &M{"gid": xa.Gid, "branch_id": branchID, "trans_type": "xa", "status": "prepared", "url": xc.CallbackURL}, "registerXaBranch", &TransOptions{})
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (rerr error) {
	xa := Xa{IDGenerator: IDGenerator{}, Gid: gid}
	data := &M{
		"gid":        gid,
		"trans_type": "xa",
	}
	rerr = CallDtm(xc.Server, data, "prepare", &TransOptions{})
	if rerr != nil {
		return
	}
	var resp *resty.Response
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := common.If(x != nil || rerr != nil, "abort", "submit").(string)
		err := CallDtm(xc.Server, data, operation, &TransOptions{})
		if rerr == nil { // 如果用户函数没有返回错误，那么返回dtm的
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	resp, rerr = xaFunc(&xa)
	rerr = CheckResponse(resp, rerr)
	return
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewBranchID()
	resp, err := common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": "action",
		}).
		Post(url)
	return resp, CheckResponse(resp, err)
}

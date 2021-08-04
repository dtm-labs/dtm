package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

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
	Gid string
	TransBase
}

// XaFromQuery construct xa info from request
func XaFromQuery(qs url.Values) (*Xa, error) {
	xa := &Xa{TransBase: *TransBaseFromQuery(qs), Gid: qs.Get("gid")}
	if xa.Gid == "" || xa.parentID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s parentid: %s", xa.Gid, xa.parentID)
	}
	return xa, nil
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
	db, err := SdbAlone(xc.Conf)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	xaID := gid + "-" + branchID
	_, err = SdbExec(db, fmt.Sprintf("xa %s '%s'", action, xaID))
	return ResultSuccess, err

}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(qs url.Values, xaFunc XaLocalFunc) (ret interface{}, rerr error) {
	xa, rerr := XaFromQuery(qs)
	if rerr != nil {
		return
	}
	xa.Dtm = xc.Server
	branchID := xa.NewBranchID()
	xaBranch := xa.Gid + "-" + branchID
	db, rerr := SdbAlone(xc.Conf)
	if rerr != nil {
		return
	}
	defer func() { db.Close() }()
	defer func() {
		x := recover()
		_, err := SdbExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
		if x == nil && rerr == nil && err == nil {
			_, err = SdbExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
		}
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = SdbExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	if rerr != nil {
		return
	}
	ret, rerr = xaFunc(db, xa)
	rerr = CheckResult(ret, rerr)
	if rerr != nil {
		return
	}
	rerr = xa.CallDtm(&M{"gid": xa.Gid, "branch_id": branchID, "trans_type": "xa", "status": "prepared", "url": xc.CallbackURL}, "registerXaBranch")
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (rerr error) {
	xa := Xa{TransBase: TransBase{IDGenerator: IDGenerator{}, Dtm: xc.Server}, Gid: gid}
	data := &M{
		"gid":        gid,
		"trans_type": "xa",
	}
	rerr = xa.CallDtm(data, "prepare")
	if rerr != nil {
		return
	}
	var resp *resty.Response
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := If(x != nil || rerr != nil, "abort", "submit").(string)
		err := xa.CallDtm(data, operation)
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
	resp, err := RestyClient.R().
		SetBody(body).
		SetQueryParams(MS{
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": "action",
		}).
		Post(url)
	return resp, CheckResponse(resp, err)
}

package dtmgrpc

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/dtmcli"
)

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *XaGrpc) (*resty.Response, error)

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *XaGrpc) (interface{}, error)

// XaRegisterCallback type of xa register callback handler
type XaRegisterCallback func(path string, xa *XaClient)

// XaClient xa client
type XaClient struct {
	Server    string
	Conf      map[string]string
	NotifyURL string
}

// XaGrpc xa transaction
type XaGrpc struct {
	dtmcli.TransData
	dtmcli.TransBase
}

// XaFromRequest construct xa info from request
func XaFromRequest(br *BusiRequest) (*XaGrpc, error) {
	xa := &XaGrpc{
		TransBase: *dtmcli.NewTransBase(br.Dtm, br.Info.BranchID),
		TransData: dtmcli.TransData{Gid: br.Info.BranchID, TransType: br.Info.TransType},
	}
	if xa.Gid == "" || br.Info.BranchID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s parentid: %s", xa.Gid, br.Info.BranchID)
	}
	return xa, nil
}

// NewXaClient construct a xa client
func NewXaClient(server string, mysqlConf map[string]string, notifyURL string, register XaRegisterCallback) (*XaClient, error) {
	xa := &XaClient{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}
	u, err := url.Parse(notifyURL)
	if err != nil {
		return nil, err
	}
	register(u.Path, xa)
	return xa, nil
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaClient) HandleCallback(gid string, branchID string, action string) (interface{}, error) {
	db, err := dtmcli.SdbAlone(xc.Conf)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	xaID := gid + "-" + branchID
	_, err = dtmcli.SdbExec(db, fmt.Sprintf("xa %s '%s'", action, xaID))
	return dtmcli.ResultSuccess, err

}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(br *BusiRequest, xaFunc XaLocalFunc) (ret interface{}, rerr error) {
	xa, rerr := XaFromRequest(br)
	if rerr != nil {
		return
	}
	xa.Dtm = xc.Server
	branchID := xa.NewBranchID()
	xaBranch := xa.Gid + "-" + branchID
	db, rerr := dtmcli.SdbAlone(xc.Conf)
	if rerr != nil {
		return
	}
	defer func() { db.Close() }()
	defer func() {
		x := recover()
		_, err := dtmcli.SdbExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
		if x == nil && rerr == nil && err == nil {
			_, err = dtmcli.SdbExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
		}
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = dtmcli.SdbExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	if rerr != nil {
		return
	}
	ret, rerr = xaFunc(db, xa)
	rerr = dtmcli.CheckResult(ret, rerr)
	if rerr != nil {
		return
	}
	rerr = xa.CallDtm(&dtmcli.M{"gid": xa.Gid, "branch_id": branchID, "trans_type": "xa", "url": xc.NotifyURL}, "registerXaBranch")
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (rerr error) {
	xa := XaGrpc{TransBase: dtmcli.TransBase{IDGenerator: dtmcli.IDGenerator{}, Dtm: xc.Server}, TransData: dtmcli.TransData{Gid: gid, TransType: "xa"}}
	rerr = xa.CallDtm(&xa.TransData, "prepare")
	if rerr != nil {
		return
	}
	var resp *resty.Response
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		operation := dtmcli.If(x != nil || rerr != nil, "abort", "submit").(string)
		err := xa.CallDtm(&xa.TransData, operation)
		if rerr == nil { // 如果用户函数没有返回错误，那么返回dtm的
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	resp, rerr = xaFunc(&xa)
	rerr = dtmcli.CheckResponse(resp, rerr)
	return
}

// CallBranch call a xa branch
func (x *XaGrpc) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewBranchID()
	resp, err := dtmcli.RestyClient.R().
		SetBody(body).
		SetQueryParams(dtmcli.MS{
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": "action",
		}).
		Post(url)
	return resp, dtmcli.CheckResponse(resp, err)
}

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
	XaClientBase
}

// Xa xa transaction
type Xa struct {
	TransBase
}

// XaFromQuery construct xa info from request
func XaFromQuery(qs url.Values) (*Xa, error) {
	xa := &Xa{TransBase: *TransBaseFromQuery(qs)}
	if xa.Gid == "" || xa.parentID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s parentid: %s", xa.Gid, xa.parentID)
	}
	return xa, nil
}

// NewXaClient construct a xa client
func NewXaClient(server string, mysqlConf map[string]string, notifyURL string, register XaRegisterCallback) (*XaClient, error) {
	xa := &XaClient{XaClientBase: XaClientBase{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}}
	u, err := url.Parse(notifyURL)
	if err != nil {
		return nil, err
	}
	register(u.Path, xa)
	return xa, nil
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaClient) HandleCallback(gid string, branchID string, action string) (interface{}, error) {
	return MapSuccess, xc.XaClientBase.HandleCallback(gid, branchID, action)
}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(qs url.Values, xaFunc XaLocalFunc) (ret interface{}, rerr error) {
	xa, rerr := XaFromQuery(qs)
	if rerr != nil {
		return
	}
	ret, rerr = xc.HandleLocalTrans(&xa.TransBase, func(db *sql.DB) (ret interface{}, rerr error) {
		ret, rerr = xaFunc(db, xa)
		rerr = CheckResult(ret, rerr)
		if rerr != nil {
			return
		}
		rerr = xa.callDtm(&M{"gid": xa.Gid, "branch_id": xa.CurrentBranchID(), "trans_type": "xa", "url": xc.XaClientBase.NotifyURL}, "registerXaBranch")
		return
	})
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (rerr error) {
	xa := Xa{TransBase: *NewTransBase(gid, "xa", xc.XaClientBase.Server, "")}
	return xc.HandleGlobalTrans(&xa.TransBase, func(action string) error {
		return xa.callDtm(xa, action)
	}, func() error {
		resp, rerr := xaFunc(&xa)
		return CheckResponse(resp, rerr)
	})
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewBranchID()
	resp, err := RestyClient.R().
		SetBody(body).
		SetQueryParams(MS{
			"dtm":         x.Dtm,
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": BranchAction,
		}).
		Post(url)
	return resp, CheckResponse(resp, err)
}

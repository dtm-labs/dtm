package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *Xa) (*resty.Response, error)

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *Xa) error

// XaRegisterCallback type of xa register callback handler
type XaRegisterCallback func(path string, xa *XaClient)

// XaClient xa client
type XaClient struct {
	dtmimp.XaClientBase
}

// Xa xa transaction
type Xa struct {
	dtmimp.TransBase
}

// XaFromQuery construct xa info from request
func XaFromQuery(qs url.Values) (*Xa, error) {
	xa := &Xa{TransBase: *dtmimp.TransBaseFromQuery(qs)}
	if xa.Gid == "" || xa.BranchID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s branchid: %s", xa.Gid, xa.BranchID)
	}
	return xa, nil
}

// NewXaClient construct a xa client
func NewXaClient(server string, mysqlConf map[string]string, notifyURL string, register XaRegisterCallback) (*XaClient, error) {
	xa := &XaClient{XaClientBase: dtmimp.XaClientBase{
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
func (xc *XaClient) XaLocalTransaction(qs url.Values, xaFunc XaLocalFunc) error {
	xa, err := XaFromQuery(qs)
	if err != nil {
		return err
	}
	return xc.HandleLocalTrans(&xa.TransBase, func(db *sql.DB) error {
		err := xaFunc(db, xa)
		if err != nil {
			return err
		}
		return dtmimp.TransRegisterBranch(&xa.TransBase, map[string]string{
			"url":       xc.XaClientBase.NotifyURL,
			"branch_id": xa.BranchID,
		}, "registerXaBranch")
	})
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (rerr error) {
	xa := Xa{TransBase: *dtmimp.NewTransBase(gid, "xa", xc.XaClientBase.Server, "")}
	return xc.HandleGlobalTrans(&xa.TransBase, func(action string) error {
		return dtmimp.TransCallDtm(&xa.TransBase, &xa, action)
	}, func() error {
		_, rerr := xaFunc(&xa)
		return rerr
	})
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewSubBranchID()
	return dtmimp.TransRequestBranch(&x.TransBase, body, branchID, BranchAction, url)
}

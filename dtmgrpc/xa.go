package dtmgrpc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yedf/dtm/dtmcli"
)

// XaGrpcGlobalFunc type of xa global function
type XaGrpcGlobalFunc func(xa *XaGrpc) error

// XaGrpcLocalFunc type of xa local function
type XaGrpcLocalFunc func(db *sql.DB, xa *XaGrpc) error

// XaGrpcClient xa client
type XaGrpcClient struct {
	Server    string
	Conf      map[string]string
	NotifyURL string
}

// XaGrpc xa transaction
type XaGrpc struct {
	dtmcli.TransBase
}

// XaGrpcFromRequest construct xa info from request
func XaGrpcFromRequest(br *BusiRequest) (*XaGrpc, error) {
	xa := &XaGrpc{
		TransBase: *dtmcli.NewTransBase(br.Info.Gid, br.Info.TransType, br.Dtm, br.Info.BranchID),
	}
	if xa.Gid == "" || br.Info.BranchID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s parentid: %s", xa.Gid, br.Info.BranchID)
	}
	return xa, nil
}

// NewXaGrpcClient construct a xa client
func NewXaGrpcClient(server string, mysqlConf map[string]string, notifyURL string) *XaGrpcClient {
	xa := &XaGrpcClient{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}
	return xa
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaGrpcClient) HandleCallback(gid string, branchID string, action string) error {
	db, err := dtmcli.SdbAlone(xc.Conf)
	if err != nil {
		return err
	}
	defer db.Close()
	xaID := gid + "-" + branchID
	_, err = dtmcli.DBExec(db, fmt.Sprintf("xa %s '%s'", action, xaID))
	return err

}

// XaLocalTransaction start a xa local transaction
func (xc *XaGrpcClient) XaLocalTransaction(br *BusiRequest, xaFunc XaGrpcLocalFunc) (rerr error) {
	xa, rerr := XaGrpcFromRequest(br)
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
		_, err := dtmcli.DBExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
		if x == nil && rerr == nil && err == nil {
			_, err = dtmcli.DBExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
		}
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = dtmcli.DBExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	if rerr != nil {
		return
	}
	rerr = xaFunc(db, xa)
	if rerr != nil {
		return
	}
	_, rerr = MustGetDtmClient(xa.Dtm).RegisterXaBranch(context.Background(), &DtmXaBranchRequest{
		Info: &BranchInfo{
			Gid:       xa.Gid,
			BranchID:  branchID,
			TransType: xa.TransType,
		},
		BusiData: "",
		Notify:   xc.NotifyURL,
	})
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaGrpcClient) XaGlobalTransaction(gid string, xaFunc XaGrpcGlobalFunc) (rerr error) {
	xa := XaGrpc{TransBase: *dtmcli.NewTransBase(gid, "xa", xc.Server, "")}
	dc := MustGetDtmClient(xa.Dtm)
	req := &DtmRequest{
		Gid:       gid,
		TransType: xa.TransType,
	}
	_, rerr = dc.Prepare(context.Background(), req)
	if rerr != nil {
		return
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		if x == nil && rerr == nil {
			_, rerr = dc.Submit(context.Background(), req)
			return
		}
		_, err := dc.Abort(context.Background(), req)
		if rerr == nil { // 如果用户函数没有返回错误，那么返回dtm的
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	rerr = xaFunc(&xa)
	return
}

// CallBranch call a xa branch
func (x *XaGrpc) CallBranch(busiData []byte, url string) (*BusiReply, error) {
	branchID := x.NewBranchID()
	server, method := GetServerAndMethod(url)
	reply := &BusiReply{}
	err := MustGetGrpcConn(server).Invoke(context.Background(), method, &BusiRequest{
		Info: &BranchInfo{
			Gid:        x.Gid,
			TransType:  x.TransType,
			BranchID:   branchID,
			BranchType: "",
		},
		Dtm:      x.Dtm,
		BusiData: busiData,
	}, reply)
	return reply, err

}

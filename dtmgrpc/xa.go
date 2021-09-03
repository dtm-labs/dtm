package dtmgrpc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yedf/dtm/dtmcli"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// XaGrpcGlobalFunc type of xa global function
type XaGrpcGlobalFunc func(xa *XaGrpc) error

// XaGrpcLocalFunc type of xa local function
type XaGrpcLocalFunc func(db *sql.DB, xa *XaGrpc) error

// XaGrpcClient xa client
type XaGrpcClient struct {
	dtmcli.XaClientBase
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
	xa := &XaGrpcClient{XaClientBase: dtmcli.XaClientBase{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}}
	return xa
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaGrpcClient) HandleCallback(gid string, branchID string, action string) error {
	return xc.XaClientBase.HandleCallback(gid, branchID, action)
}

// XaLocalTransaction start a xa local transaction
func (xc *XaGrpcClient) XaLocalTransaction(br *BusiRequest, xaFunc XaGrpcLocalFunc) (rerr error) {
	xa, rerr := XaGrpcFromRequest(br)
	if rerr != nil {
		return
	}
	_, rerr = xc.HandleLocalTrans(&xa.TransBase, func(db *sql.DB) (interface{}, error) {
		rerr := xaFunc(db, xa)
		if rerr != nil {
			return nil, rerr
		}
		_, rerr = MustGetDtmClient(xa.Dtm).RegisterXaBranch(context.Background(), &DtmXaBranchRequest{
			Info: &BranchInfo{
				Gid:       xa.Gid,
				BranchID:  xa.CurrentBranchID(),
				TransType: xa.TransType,
			},
			BusiData: "",
			Notify:   xc.NotifyURL,
		})
		return nil, rerr
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
	return xc.HandleGlobalTrans(&xa.TransBase, func(action string) error {
		f := map[string]func(context.Context, *DtmRequest, ...grpc.CallOption) (*emptypb.Empty, error){
			"prepare": dc.Prepare,
			"submit":  dc.Submit,
			"abort":   dc.Abort,
		}[action]
		_, err := f(context.Background(), req)
		return err
	}, func() error {
		return xaFunc(&xa)
	})
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

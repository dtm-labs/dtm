package dtmgrpc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// XaGrpcGlobalFunc type of xa global function
type XaGrpcGlobalFunc func(xa *XaGrpc) error

// XaGrpcLocalFunc type of xa local function
type XaGrpcLocalFunc func(db *sql.DB, xa *XaGrpc) error

// XaGrpcClient xa client
type XaGrpcClient struct {
	dtmimp.XaClientBase
}

// XaGrpc xa transaction
type XaGrpc struct {
	dtmimp.TransBase
}

// XaGrpcFromRequest construct xa info from request
func XaGrpcFromRequest(ctx context.Context) (*XaGrpc, error) {
	xa := &XaGrpc{
		TransBase: *dtmgimp.TransBaseFromGrpc(ctx),
	}
	if xa.Gid == "" || xa.BranchID == "" {
		return nil, fmt.Errorf("bad xa info: gid: %s branchid: %s", xa.Gid, xa.BranchID)
	}
	return xa, nil
}

// NewXaGrpcClient construct a xa client
func NewXaGrpcClient(server string, mysqlConf map[string]string, notifyURL string) *XaGrpcClient {
	xa := &XaGrpcClient{XaClientBase: dtmimp.XaClientBase{
		Server:    server,
		Conf:      mysqlConf,
		NotifyURL: notifyURL,
	}}
	return xa
}

// HandleCallback 处理commit/rollback的回调
func (xc *XaGrpcClient) HandleCallback(ctx context.Context) (*emptypb.Empty, error) {
	tb := dtmgimp.TransBaseFromGrpc(ctx)
	return &emptypb.Empty{}, xc.XaClientBase.HandleCallback(tb.Gid, tb.BranchID, tb.Op)
}

// XaLocalTransaction start a xa local transaction
func (xc *XaGrpcClient) XaLocalTransaction(ctx context.Context, msg proto.Message, xaFunc XaGrpcLocalFunc) error {
	xa, err := XaGrpcFromRequest(ctx)
	if err != nil {
		return err
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return xc.HandleLocalTrans(&xa.TransBase, func(db *sql.DB) error {
		err := xaFunc(db, xa)
		if err != nil {
			return err
		}
		_, err = dtmgimp.MustGetDtmClient(xa.Dtm).RegisterBranch(context.Background(), &dtmgimp.DtmBranchRequest{
			Gid:         xa.Gid,
			BranchID:    xa.BranchID,
			TransType:   xa.TransType,
			BusiPayload: data,
			Data:        map[string]string{"url": xc.NotifyURL},
		})
		return err
	})
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaGrpcClient) XaGlobalTransaction(gid string, xaFunc XaGrpcGlobalFunc) error {
	xa := XaGrpc{TransBase: *dtmimp.NewTransBase(gid, "xa", xc.Server, "")}
	dc := dtmgimp.MustGetDtmClient(xa.Dtm)
	req := &dtmgimp.DtmRequest{
		Gid:       gid,
		TransType: xa.TransType,
	}
	return xc.HandleGlobalTrans(&xa.TransBase, func(action string) error {
		f := map[string]func(context.Context, *dtmgimp.DtmRequest, ...grpc.CallOption) (*emptypb.Empty, error){
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
func (x *XaGrpc) CallBranch(msg proto.Message, url string, reply interface{}) error {
	server, method := dtmgimp.GetServerAndMethod(url)
	err := dtmgimp.MustGetGrpcConn(server, false).Invoke(
		dtmgimp.TransInfo2Ctx(x.Gid, x.TransType, x.NewSubBranchID(), "action", x.Dtm), method, msg, reply)
	return err

}

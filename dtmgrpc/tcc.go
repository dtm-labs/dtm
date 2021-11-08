package dtmgrpc

import (
	context "context"
	"fmt"

	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/protobuf/proto"
)

// TccGrpc struct of tcc
type TccGrpc struct {
	dtmimp.TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *TccGrpc) error

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm服务器地址
// gid 全局事务id
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	tcc := &TccGrpc{TransBase: *dtmimp.NewTransBase(gid, "tcc", dtm, "")}
	dc := dtmgimp.MustGetDtmClient(tcc.Dtm)
	dr := &dtmgimp.DtmRequest{
		Gid:       tcc.Gid,
		TransType: tcc.TransType,
	}
	_, rerr = dc.Prepare(context.Background(), dr)
	if rerr != nil {
		return rerr
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		x := recover()
		if x == nil && rerr == nil {
			_, rerr = dc.Submit(context.Background(), dr)
			return
		}
		_, err := dc.Abort(context.Background(), dr)
		if rerr == nil {
			rerr = err
		}
		if x != nil {
			panic(x)
		}
	}()
	return tccFunc(tcc)
}

// TccFromGrpc tcc from request info
func TccFromGrpc(ctx context.Context) (*TccGrpc, error) {
	tcc := &TccGrpc{
		TransBase: *dtmgimp.TransBaseFromGrpc(ctx),
	}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s branchid: %s", tcc.Dtm, tcc.Gid, tcc.BranchID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
func (t *TccGrpc) CallBranch(busiMsg proto.Message, tryURL string, confirmURL string, cancelURL string, reply interface{}) error {
	branchID := t.NewSubBranchID()
	bd, err := proto.Marshal(busiMsg)
	_, err = dtmgimp.MustGetDtmClient(t.Dtm).RegisterBranch(context.Background(), &dtmgimp.DtmBranchRequest{
		Gid:         t.Gid,
		TransType:   t.TransType,
		BranchID:    branchID,
		BusiPayload: bd,
		Data:        map[string]string{"confirm": confirmURL, "cancel": cancelURL},
	})
	if err != nil {
		return err
	}
	server, method := dtmgimp.GetServerAndMethod(tryURL)
	return dtmgimp.MustGetGrpcConn(server, false).Invoke(
		dtmgimp.TransInfo2Ctx(t.Gid, t.TransType, branchID, "try", t.Dtm), method, busiMsg, reply)
}

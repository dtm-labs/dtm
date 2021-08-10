package dtmgrpc

import (
	context "context"
	"fmt"

	"github.com/yedf/dtm/dtmcli"
)

// TccGrpc struct of tcc
type TccGrpc struct {
	dtmcli.TransData
	dtmcli.TransBase
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *TccGrpc) error

// TccGlobalTransaction begin a tcc global transaction
// dtm dtm服务器地址
// gid 全局事务id
// tccFunc tcc事务函数，里面会定义全局事务的分支
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	tcc := &TccGrpc{TransBase: dtmcli.TransBase{Dtm: dtm}, TransData: dtmcli.TransData{Gid: gid, TransType: "tcc"}}
	dc := MustGetDtmClient(tcc.Dtm)
	dr := &DtmRequest{
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
		} else {
			_, err := dc.Abort(context.Background(), dr)
			if rerr == nil {
				rerr = err
			}
			if x != nil {
				panic(x)
			}

		}
	}()
	return tccFunc(tcc)
}

// TccFromRequest tcc from request info
func TccFromRequest(br *BusiRequest) (*TccGrpc, error) {
	tcc := &TccGrpc{
		TransBase: *dtmcli.NewTransBase(br.Dtm, br.Info.BranchID),
		TransData: dtmcli.TransData{Gid: br.Info.BranchID, TransType: br.Info.TransType},
	}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s parentID: %s", tcc.Dtm, tcc.Gid, br.Info.BranchID)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
// 函数首先注册子事务的所有分支，成功后调用try分支，返回try分支的调用结果
func (t *TccGrpc) CallBranch(busiData []byte, tryURL string, confirmURL string, cancelURL string) (*BusiReply, error) {
	branchID := t.NewBranchID()
	_, err := MustGetDtmClient(t.Dtm).RegisterTccBranch(context.Background(), &DtmTccBranchRequest{
		Info: &BranchInfo{
			Gid:       t.Gid,
			TransType: t.TransType,
			BranchID:  branchID,
		},
		BusiData: string(busiData),
		Try:      tryURL,
		Confirm:  confirmURL,
		Cancel:   cancelURL,
	})
	if err != nil {
		return nil, err
	}
	server, method := GetServerAndMethod(tryURL)
	reply := &BusiReply{}
	err = MustGetGrpcConn(server).Invoke(context.Background(), method, &BusiRequest{
		Info: &BranchInfo{
			Gid:        t.Gid,
			TransType:  t.TransType,
			BranchID:   branchID,
			BranchType: "try",
		},
		BusiData: busiData,
	}, reply)
	return reply, err
}

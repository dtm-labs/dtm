package workflow

import (
	"context"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (wf *Workflow) getProgress() ([]*dtmgpb.DtmProgress, error) {
	if wf.Protocol == dtmimp.ProtocolGRPC {
		var reply dtmgpb.DtmProgressesReply
		err := dtmgimp.MustGetGrpcConn(wf.Dtm, false).Invoke(wf.Context, "/dtmgimp.Dtm/PrepareWorkflow",
			dtmgimp.GetDtmRequest(wf.TransBase), &reply)
		if err == nil {
			return reply.Progresses, nil
		}
		return nil, err
	}
	resp, err := dtmcli.GetRestyClient().R().SetBody(wf.TransBase).Post(wf.Dtm + "/prepareWorkflow")
	var reply dtmgpb.DtmProgressesReply
	if err == nil {
		dtmimp.MustUnmarshal(resp.Body(), &reply)
	}
	return reply.Progresses, err
}

func (wf *Workflow) submit(err error) error {
	status := wfErrorToStatus(err)
	reason := ""
	if err != nil {
		reason = err.Error()
	}
	if wf.Protocol == dtmimp.ProtocolHTTP {
		m := map[string]interface{}{
			"gid":        wf.Gid,
			"trans_type": wf.TransType,
			"req_extra": map[string]string{
				"status":          status,
				"rollback_reason": reason,
			},
		}
		_, err := dtmimp.TransCallDtmExt(wf.TransBase, m, "submit")
		return err
	}
	req := dtmgimp.GetDtmRequest(wf.TransBase)
	req.ReqExtra = map[string]string{
		"status":          status,
		"rollback_reason": reason,
	}
	reply := emptypb.Empty{}
	return dtmgimp.MustGetGrpcConn(wf.Dtm, false).Invoke(wf.Context, "/dtmgimp.Dtm/"+"Submit", req, &reply)
}

func (wf *Workflow) registerBranch(res []byte, branchID string, op string, status string) error {
	if wf.Protocol == dtmimp.ProtocolHTTP {
		return dtmimp.TransRegisterBranch(wf.TransBase, map[string]string{
			"data":      string(res),
			"branch_id": branchID,
			"op":        op,
			"status":    status,
		}, "registerBranch")
	}
	_, err := dtmgimp.MustGetDtmClient(wf.Dtm).RegisterBranch(context.Background(), &dtmgpb.DtmBranchRequest{
		Gid:         wf.Gid,
		TransType:   wf.TransType,
		BranchID:    branchID,
		BusiPayload: res,
		Data:        map[string]string{"status": status, "op": op},
	})
	return err
}

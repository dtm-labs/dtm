package workflow

import (
	"context"
	"encoding/base64"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmcli/logger"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (wf *Workflow) getProgress() (*dtmgpb.DtmProgressesReply, error) {
	if wf.Protocol == dtmimp.ProtocolGRPC {
		var reply dtmgpb.DtmProgressesReply
		err := dtmgimp.MustGetGrpcConn(wf.Dtm, false).Invoke(wf.Context, "/dtmgimp.Dtm/PrepareWorkflow",
			dtmgimp.GetDtmRequest(wf.TransBase), &reply)
		return &reply, err
	}
	resp, err := dtmcli.GetRestyClient().R().SetBody(wf.TransBase).Post(wf.Dtm + "/prepareWorkflow")
	var reply dtmgpb.DtmProgressesReply
	if err == nil {
		uo := protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}
		err = uo.Unmarshal(resp.Body(), &reply)
	}
	logger.Infof(resp.String())
	return &reply, err
}

func (wf *Workflow) submit(result []byte, err error) error {
	status := wfErrorToStatus(err)
	reason := ""
	if err != nil {
		reason = err.Error()
	}
	extra := map[string]string{
		"status":          status,
		"rollback_reason": reason,
		"result":          base64.StdEncoding.EncodeToString(result),
	}
	if wf.Protocol == dtmimp.ProtocolHTTP {
		m := map[string]interface{}{
			"gid":        wf.Gid,
			"trans_type": wf.TransType,
			"req_extra":  extra,
		}
		_, err := dtmimp.TransCallDtmExt(wf.TransBase, m, "submit")
		return err
	}
	req := dtmgimp.GetDtmRequest(wf.TransBase)
	req.ReqExtra = extra
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

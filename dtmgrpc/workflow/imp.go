package workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/go-resty/resty/v2"
)

type workflowImp struct {
	restyClient          *resty.Client //nolint
	idGen                dtmimp.BranchIDGen
	currentBranch        string                 //nolint
	currentActionAdded   bool                   //nolint
	currentCommitAdded   bool                   //nolint
	currentRollbackAdded bool                   //nolint
	currentRollbackItem  *workflowPhase2Item    // nolint
	progresses           map[string]*stepResult //nolint
	currentOp            string
	succeededOps         []workflowPhase2Item
	failedOps            []workflowPhase2Item
}

type workflowPhase2Item struct {
	branchID, op string
	fn           WfPhase2Func
}

func (wf *Workflow) loadProgresses() error {
	progresses, err := wf.getProgress()
	if err == nil {
		wf.progresses = map[string]*stepResult{}
		for _, p := range progresses {
			sr := &stepResult{
				Status: p.Status,
				Data:   p.BinData,
			}
			if sr.Status == dtmcli.StatusFailed {
				sr.Error = fmt.Errorf("%s. %w", string(p.BinData), dtmcli.ErrFailure)
			}
			wf.progresses[p.BranchID+"-"+p.Op] = sr
		}
	}
	return err
}

type wfMeta struct{}

func (w *workflowFactory) newWorkflow(name string, gid string, data []byte) *Workflow {
	wf := &Workflow{
		TransBase: dtmimp.NewTransBase(gid, "workflow", "not inited", ""),
		Name:      name,
		workflowImp: workflowImp{
			idGen:        dtmimp.BranchIDGen{},
			succeededOps: []workflowPhase2Item{},
			failedOps:    []workflowPhase2Item{},
			currentOp:    dtmimp.OpAction,
		},
	}
	wf.Protocol = w.protocol
	if w.protocol == dtmimp.ProtocolGRPC {
		wf.Dtm = w.grpcDtm
		wf.QueryPrepared = w.grpcCallback
	} else {
		wf.Dtm = w.httpDtm
		wf.QueryPrepared = w.httpCallback
	}
	wf.CustomData = dtmimp.MustMarshalString(map[string]interface{}{
		"name": wf.Name,
		"data": data,
	})
	wf.Context = context.WithValue(wf.Context, wfMeta{}, wf)
	wf.Options.HTTPResp2DtmError = HTTPResp2DtmError
	wf.Options.GRPCError2DtmError = GrpcError2DtmError
	wf.initRestyClient()
	return wf
}

func (wf *Workflow) initRestyClient() {
	wf.restyClient = resty.New()
	wf.restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetQueryParams(map[string]string{
			"gid":        wf.Gid,
			"trans_type": wf.TransType,
			"branch_id":  wf.currentBranch,
			"op":         dtmimp.OpAction,
		})
		err := dtmimp.BeforeRequest(c, r)
		return err
	})
	old := wf.restyClient.GetClient().Transport
	wf.restyClient.GetClient().Transport = newRoundTripper(old, wf)
	wf.restyClient.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		return dtmimp.AfterResponse(c, r)
	})
}

func (wf *Workflow) process(handler WfFunc, data []byte) (err error) {
	err = wf.loadProgresses()
	if err == nil {
		err = handler(wf, data)
		err = wf.Options.GRPCError2DtmError(err)
		if err != nil && !errors.Is(err, dtmcli.ErrFailure) {
			return err
		}
		err = wf.processPhase2(err)
	}
	if err == nil || errors.Is(err, dtmcli.ErrFailure) {
		err1 := wf.submit(wfErrorToStatus(err))
		if err1 != nil {
			return err1
		}
	}
	return err

}

func (wf *Workflow) saveResult(branchID string, op string, sr *stepResult) error {
	if sr.Status != "" {
		err := wf.registerBranch(sr.Data, branchID, op, sr.Status)
		if err != nil {
			return err
		}
	}
	return sr.Error
}

func (wf *Workflow) processPhase2(err error) error {
	ops := wf.succeededOps
	if err == nil {
		wf.currentOp = dtmimp.OpCommit
	} else {
		wf.currentOp = dtmimp.OpRollback
		ops = wf.failedOps
	}
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]

		err1 := wf.callPhase2(op.branchID, op.fn)
		if err1 != nil {
			return err1
		}
	}
	return err
}

func (wf *Workflow) callPhase2(branchID string, fn WfPhase2Func) error {
	wf.currentBranch = branchID
	r := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		err := fn(bb)
		dtmimp.PanicIf(errors.Is(err, dtmcli.ErrFailure), errors.New("should not return ErrFail in phase2"))
		return wf.stepResultFromLocal(nil, err)
	})
	_, err := wf.stepResultToLocal(r)
	return err
}

func (wf *Workflow) recordedDo(fn func(bb *dtmcli.BranchBarrier) *stepResult) *stepResult {
	sr := wf.recordedDoInner(fn)
	if wf.currentRollbackItem != nil && (sr.Status == dtmcli.StatusSucceed || sr.Status == dtmcli.StatusFailed && wf.Options.CompensateErrorBranch) {
		wf.failedOps = append(wf.failedOps, *wf.currentRollbackItem)
	}
	wf.currentRollbackItem = nil
	return sr
}

func (wf *Workflow) recordedDoInner(fn func(bb *dtmcli.BranchBarrier) *stepResult) *stepResult {
	branchID := wf.currentBranch
	if wf.currentOp == dtmimp.OpAction {
		dtmimp.PanicIf(wf.currentActionAdded, fmt.Errorf("one branch can have only on action"))
		wf.currentActionAdded = true
	}
	r := wf.getStepResult()
	if r != nil {
		logger.Debugf("progress restored: %s %s %v %s %s", branchID, wf.currentOp, r.Error, r.Status, r.Data)
		return r
	}
	bb := &dtmcli.BranchBarrier{
		TransType: wf.TransType,
		Gid:       wf.Gid,
		BranchID:  branchID,
		Op:        wf.currentOp,
	}
	r = fn(bb)
	err := wf.saveResult(branchID, wf.currentOp, r)
	if err != nil {
		r = wf.stepResultFromLocal(nil, err)
	}
	return r
}

func (wf *Workflow) getStepResult() *stepResult {
	logger.Debugf("getStepResult: %s %v", wf.currentBranch+"-"+wf.currentOp, wf.progresses[wf.currentBranch+"-"+wf.currentOp])
	return wf.progresses[wf.currentBranch+"-"+wf.currentOp]
}

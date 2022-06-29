package workflow

import (
	"context"
	"errors"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/go-resty/resty/v2"
)

type workflowImp struct {
	restyClient   *resty.Client
	idGen         dtmimp.BranchIDGen
	currentBranch string
	progresses    map[string]*stepResult
	currentOp     string
	succeededOps  []workflowPhase2Item
	failedOps     []workflowPhase2Item
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
			wf.progresses[p.BranchID+"-"+p.Op] = &stepResult{
				Status: p.Status,
				Data:   p.BinData,
			}
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
	wf.newBranch()
	wf.CustomData = dtmimp.MustMarshalString(map[string]interface{}{
		"name": wf.Name,
		"data": data,
	})
	wf.Context = context.WithValue(wf.Context, wfMeta{}, wf)
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
	wf.restyClient.GetClient().Transport = NewRoundTripper(old, wf)
	wf.restyClient.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		err := dtmimp.AfterResponse(c, r)
		if err == nil && !wf.Options.DisalbeAutoError {
			err = dtmimp.RespAsErrorCompatible(r) // check for dtm error
		}
		return err
	})
}

func (wf *Workflow) process(handler WfFunc, data []byte) (err error) {
	err = wf.prepare()
	if err == nil {
		err = wf.loadProgresses()
	}
	if err == nil {
		err = handler(wf, data)
		err = dtmgrpc.GrpcError2DtmError(err)
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
	if sr.Status == "" {
		return sr.Error
	}
	return wf.registerBranch(sr.Data, branchID, op, sr.Status)
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

		err1 := wf.callPhase2(op.branchID, op.op, op.fn)
		if err1 != nil {
			return err1
		}
	}
	return err
}

func (wf *Workflow) callPhase2(branchID string, op string, fn WfPhase2Func) error {
	wf.currentBranch = branchID
	r := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		err := fn(bb)
		if errors.Is(err, dtmcli.ErrFailure) {
			panic("should not return ErrFail in phase2")
		}
		return stepResultFromLocal(nil, err)
	})
	_, err := stepResultToLocal(r)
	return err
}

func (wf *Workflow) recordedDo(fn func(bb *dtmcli.BranchBarrier) *stepResult) *stepResult {
	branchID := wf.currentBranch
	r := wf.getStepResult()
	if wf.currentOp == dtmimp.OpAction { // for action steps, an action will start a new branch
		wf.newBranch()
	}
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
		r = stepResultFromLocal(nil, err)
	}
	return r
}

func (wf *Workflow) newBranch() {
	wf.idGen.NewSubBranchID()
	wf.currentBranch = wf.idGen.CurrentSubBranchID()
}

func (wf *Workflow) getStepResult() *stepResult {
	logger.Debugf("getStepResult: %s %v", wf.currentBranch+"-"+wf.currentOp, wf.progresses[wf.currentBranch+"-"+wf.currentOp])
	return wf.progresses[wf.currentBranch+"-"+wf.currentOp]
}

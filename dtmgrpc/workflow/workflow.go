package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/dtmgrpc/workflow/wfpb"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
)

// InitHTTP will init Workflow engine to use http
// param httpDtm specify the dtm address
// param callback specify the url for dtm to callback if a workflow timeout
func InitHTTP(httpDtm string, callback string) {
	defaultFac.protocol = dtmimp.ProtocolHTTP
	defaultFac.httpDtm = httpDtm
	defaultFac.httpCallback = callback
}

// InitGrpc will init Workflow engine to use grpc
// param dtm specify the dtm address
// param clientHost specify the client host for dtm to callback if a workflow timeout
// param grpcServer specify the grpc server
func InitGrpc(grpcDtm string, clientHost string, grpcServer *grpc.Server) {
	defaultFac.protocol = dtmimp.ProtocolGRPC
	defaultFac.grpcDtm = grpcDtm
	wfpb.RegisterWorkflowServer(grpcServer, &workflowServer{})
	defaultFac.grpcCallback = clientHost + "/workflow.Workflow/Execute"
}

// SetProtocolForTest change protocol directly. only used by test
func SetProtocolForTest(protocol string) {
	defaultFac.protocol = protocol
}

// Register will register a workflow with the specified name
func Register(name string, handler WfFunc, custom ...func(wf *Workflow)) error {
	return defaultFac.register(name, handler, custom...)
}

// Execute will execute a workflow with the gid and specified params
// if the workflow with the gid does not exist, then create a new workflow and execute it
// if the workflow with the gid exists, resume to execute it
func Execute(name string, gid string, data []byte) error {
	return defaultFac.execute(name, gid, data)
}

// ExecuteByQS is like Execute, but name and gid will be obtained from qs
func ExecuteByQS(qs url.Values, body []byte) error {
	return defaultFac.executeByQS(qs, body)
}

// Options is for specifying workflow options
type Options struct {

	// Default == HTTPResp2DtmError : Code 409 => ErrFailure; Code 425 => ErrOngoing
	HTTPResp2DtmError func(*http.Response) ([]byte, error)

	// Default == GrpcError2DtmError: Code Aborted => ErrFailure; Code FailedPrecondition => ErrOngoing
	GRPCError2DtmError func(error) error

	// This Option specify whether a branch returning ErrFailure should be compensated on rollback.
	// for most idempotent branches, no compensation is needed.
	// But for a timeout request, the caller cannot know where the request is successful, so the compensation should be called
	CompensateErrorBranch bool
}

// Workflow is the type for a workflow
type Workflow struct {
	// The name of the workflow
	Name    string
	Options Options
	*dtmimp.TransBase
	workflowImp
}

type wfItem struct {
	fn     WfFunc
	custom []func(*Workflow)
}

// WfFunc is the type for workflow function
type WfFunc func(wf *Workflow, data []byte) error

// WfPhase2Func is the type for phase 2 function
// param bb is a BranchBarrier, which is introduced by http://d.dtm.pub/practice/barrier.html
type WfPhase2Func func(bb *dtmcli.BranchBarrier) error

// NewRequest return a new resty request, whose progress will be recorded
func (wf *Workflow) NewRequest() *resty.Request {
	return wf.restyClient.R().SetContext(wf.Context)
}

// NewBranch will start a new branch transaction
func (wf *Workflow) NewBranch() *Workflow {
	dtmimp.PanicIf(wf.currentOp != dtmimp.OpAction, fmt.Errorf("should not call NewBranch() in Branch callbacks"))
	wf.idGen.NewSubBranchID()
	wf.currentBranch = wf.idGen.CurrentSubBranchID()
	wf.currentActionAdded = false
	wf.currentCommitAdded = false
	wf.currentRollbackAdded = false
	return wf
}

// NewBranchCtx will call NewBranch and return a workflow context
func (wf *Workflow) NewBranchCtx() context.Context {
	return wf.NewBranch().Context
}

// OnRollback will set the callback for current branch when rollback happen.
// If you are writing a saga transaction, then you should  write the compensation here
// If you are writing a tcc transaction, then you should write the cancel operation here
func (wf *Workflow) OnRollback(compensate WfPhase2Func) *Workflow {
	branchID := wf.currentBranch
	dtmimp.PanicIf(wf.currentRollbackAdded, fmt.Errorf("one branch can only add one rollback callback"))
	wf.currentRollbackAdded = true
	item := workflowPhase2Item{
		branchID: branchID,
		op:       dtmimp.OpRollback,
		fn:       compensate,
	}
	wf.currentRollbackItem = &item
	return wf
}

// OnCommit will will set the callback for current branch when commit happen.
// If you are writing a tcc transaction, then you should write the confirm operation here
func (wf *Workflow) OnCommit(fn WfPhase2Func) *Workflow {
	branchID := wf.currentBranch
	dtmimp.PanicIf(wf.currentCommitAdded, fmt.Errorf("one branch can only add one commit callback"))
	wf.currentCommitAdded = true
	wf.failedOps = append(wf.succeededOps, workflowPhase2Item{
		branchID: branchID,
		op:       dtmimp.OpCommit,
		fn:       fn,
	})
	return wf
}

// OnFinish will both set the callback for OnCommit and OnRollback
func (wf *Workflow) OnFinish(fn func(bb *dtmcli.BranchBarrier, isRollback bool) error) *Workflow {
	return wf.OnCommit(func(bb *dtmcli.BranchBarrier) error {
		return fn(bb, false)
	}).OnRollback(func(bb *dtmcli.BranchBarrier) error {
		return fn(bb, true)
	})
}

// Do will do an action which will be recored
func (wf *Workflow) Do(fn func(bb *dtmcli.BranchBarrier) ([]byte, error)) ([]byte, error) {
	res := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		r, e := fn(bb)
		return wf.stepResultFromLocal(r, e)
	})
	return wf.stepResultToLocal(res)
}

// DoXa will begin a local xa transaction
// after the return of workflow function, xa commit/rollback will be called
func (wf *Workflow) DoXa(dbConf dtmcli.DBConf, fn func(db *sql.DB) ([]byte, error)) ([]byte, error) {
	branchID := wf.currentBranch
	res := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		sBusi := "business"
		k := bb.BranchID + "-" + sBusi
		if wf.progresses[k] != nil {
			return &stepResult{
				Error: fmt.Errorf("error occur at prepare, not resumable, to rollback. %w", dtmcli.ErrFailure),
			}
		}
		sr := &stepResult{}
		wf.TransBase.BranchID = branchID
		wf.TransBase.Op = sBusi
		err := dtmimp.XaHandleLocalTrans(wf.TransBase, dbConf, func(d *sql.DB) error {
			r, e := fn(d)
			sr.Data = r
			if e == nil {
				e = wf.saveResult(branchID, sBusi, &stepResult{Status: dtmcli.StatusSucceed})
			}
			return e
		})
		sr.Error = err
		sr.Status = wfErrorToStatus(err)
		return sr
	})
	phase2 := func(bb *dtmcli.BranchBarrier) error {
		return dtmimp.XaHandlePhase2(bb.Gid, dbConf, bb.BranchID, bb.Op)
	}
	wf.succeededOps = append(wf.succeededOps, workflowPhase2Item{
		branchID: branchID,
		op:       dtmimp.OpCommit,
		fn:       phase2,
	})
	wf.failedOps = append(wf.failedOps, workflowPhase2Item{
		branchID: branchID,
		op:       dtmimp.OpRollback,
		fn:       phase2,
	})
	return res.Data, res.Error
}

// Interceptor is the middleware for workflow to capture grpc call result
func Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger.Debugf("grpc client calling: %s%s %v", cc.Target(), method, dtmimp.MustMarshalString(req))
	wf := ctx.Value(wfMeta{}).(*Workflow)

	origin := func() error {
		ctx1 := dtmgimp.TransInfo2Ctx(ctx, wf.Gid, wf.TransType, wf.currentBranch, wf.currentOp, wf.Dtm)
		err := invoker(ctx1, method, req, reply, cc, opts...)
		res := fmt.Sprintf("grpc client called: %s%s %s result: %s err: %v",
			cc.Target(), method, dtmimp.MustMarshalString(req), dtmimp.MustMarshalString(reply), err)
		if err != nil {
			logger.Errorf("%s", res)
		} else {
			logger.Debugf("%s", res)
		}
		return err
	}
	if wf.currentOp != dtmimp.OpAction {
		return origin()
	}
	sr := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		err := origin()
		return wf.stepResultFromGrpc(reply, err)
	})
	return wf.stepResultToGrpc(sr, reply)
}

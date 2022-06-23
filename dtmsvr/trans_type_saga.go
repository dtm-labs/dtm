package dtmsvr

import (
	"errors"
	"fmt"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
)

type transSagaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("saga", func(trans *TransGlobal) transProcessor {
		return &transSagaProcessor{TransGlobal: trans}
	})
}

func (t *transSagaProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	for i, step := range t.Steps {
		branch := fmt.Sprintf("%02d", i+1)
		for _, op := range []string{dtmimp.OpCompensate, dtmimp.OpAction} {
			branches = append(branches, TransBranch{
				Gid:      t.Gid,
				BranchID: branch,
				BinData:  t.BinPayloads[i],
				URL:      step[op],
				Op:       op,
				Status:   dtmcli.StatusPrepared,
			})
		}
	}
	return branches
}

type cSagaCustom struct {
	Orders     map[int][]int `json:"orders"`
	Concurrent bool          `json:"concurrent"`
	cOrders    map[int][]int
}

type branchResult struct {
	index   int
	status  string
	started bool
	op      string
	err     error
}

func (t *transSagaProcessor) ProcessOnce(branches []TransBranch) error {
	// when saga tasks is fetched, it always need to process
	logger.Debugf("status: %s timeout: %t", t.Status, t.isTimeout())
	if t.Status == dtmcli.StatusSubmitted && t.isTimeout() {
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(fmt.Sprintf("Timeout after %d seconds", t.TimeoutToFail)))
	}
	n := len(branches)

	csc := cSagaCustom{Orders: map[int][]int{}, cOrders: map[int][]int{}}
	if t.CustomData != "" {
		dtmimp.MustUnmarshalString(t.CustomData, &csc)
		for k, v := range csc.Orders {
			for _, b := range v {
				csc.cOrders[b] = append(csc.cOrders[b], k)
			}
		}
	}
	if csc.Concurrent || t.TimeoutToFail > 0 { // when saga is not normal, update branch sync
		t.updateBranchSync = true
	}
	// resultStats
	var rsAToStart, rsAStarted, rsADone, rsAFailed, rsASucceed, rsCToStart, rsCDone, rsCSucceed int
	var failureError error
	branchResults := make([]branchResult, n) // save the branch result
	for i := 0; i < n; i++ {
		b := branches[i]
		if b.Op == dtmimp.OpAction {
			if b.Status == dtmcli.StatusPrepared {
				rsAToStart++
			} else if b.Status == dtmcli.StatusFailed {
				rsAFailed++
			}
		}
		branchResults[i] = branchResult{index: i, status: branches[i].Status, op: branches[i].Op}
	}
	shouldRun := func(current int) bool {
		// if !csc.Concurrent，then check the branch in previous step is succeed
		if !csc.Concurrent && current >= 2 && branchResults[current-2].status != dtmcli.StatusSucceed {
			return false
		}
		// if csc.concurrent, then check the Orders. origin one step correspond to 2 step in dtmsvr
		for _, pre := range csc.Orders[current/2] {
			if branchResults[pre*2+1].status != dtmcli.StatusSucceed {
				return false
			}
		}
		return true
	}
	shouldRollback := func(current int) bool {
		rollbacked := func(i int) bool {
			// current compensate op rollbacked or related action still prepared
			return branchResults[i].status == dtmcli.StatusSucceed || branchResults[i+1].status == dtmcli.StatusPrepared
		}
		if rollbacked(current) {
			return false
		}
		// if !csc.Concurrent，then check the branch in next step is rollbacked
		if !csc.Concurrent && current < n-2 && !rollbacked(current+2) {
			return false
		}
		// if csc.concurrent, then check the cOrders. origin one step correspond to 2 step in dtmsvr
		for _, next := range csc.cOrders[current/2] {
			if !rollbacked(2 * next) {
				return false
			}
		}
		return true
	}
	resultChan := make(chan branchResult, n)
	asyncExecBranch := func(i int) {
		var err error
		defer func() {
			if x := recover(); x != nil {
				err = dtmimp.AsError(x)
			}
			resultChan <- branchResult{index: i, status: branches[i].Status, op: branches[i].Op, err: branches[i].Error}
			if err != nil && !errors.Is(err, dtmcli.ErrOngoing) {
				logger.Errorf("exec branch %s %s %s error: %v", branches[i].BranchID, branches[i].Op, branches[i].URL, err)
			}
		}()
		err = t.execBranch(&branches[i], i)
	}
	pickToRunActions := func() []int {
		toRun := []int{}
		for current := 1; current < n; current += 2 {
			br := &branchResults[current]
			if !br.started && br.status == dtmcli.StatusPrepared && shouldRun(current) {
				toRun = append(toRun, current)
			}
		}
		logger.Debugf("toRun picked for action is: %v branchResults: %v compensate orders: %v", toRun, branchResults, csc.cOrders)
		return toRun
	}
	pickToRunCompensates := func() []int {
		toRun := []int{}
		for current := n - 2; current >= 0; current -= 2 {
			br := &branchResults[current]
			if !br.started && br.status == dtmcli.StatusPrepared && shouldRollback(current) {
				toRun = append(toRun, current)
			}
		}
		logger.Debugf("toRun picked for compensate is: %v branchResults: %v compensate orders: %v", toRun, branchResults, csc.cOrders)
		return toRun
	}
	runBranches := func(toRun []int) {
		for _, b := range toRun {
			branchResults[b].started = true
			if branchResults[b].op == dtmimp.OpAction {
				rsAStarted++
			}
			go asyncExecBranch(b)
		}
	}
	waitDoneOnce := func() {
		select {
		case r := <-resultChan:
			br := &branchResults[r.index]
			br.status = r.status
			if r.op == dtmimp.OpAction {
				rsADone++
				if r.status == dtmcli.StatusFailed {
					rsAFailed++
					failureError = r.err
				} else if r.status == dtmcli.StatusSucceed {
					rsASucceed++
				}
			} else {
				rsCDone++
				if r.status == dtmcli.StatusSucceed {
					rsCSucceed++
				}
			}
			logger.Debugf("branch done: %v", r)
		case <-time.After(time.Second * 3):
			logger.Debugf("wait once for done")
		}
	}
	prepareToCompensate := func() {
		toRun := pickToRunActions()
		for _, b := range toRun { // flag started
			branchResults[b].started = true
		}
		for i := 1; i < len(branchResults); i += 2 {
			// these branches may have run. so flag them to status succeed, then run the corresponding
			// compensate
			if branchResults[i].started && branchResults[i].status == dtmcli.StatusPrepared {
				branchResults[i].status = dtmcli.StatusSucceed
			}
		}
		for i, b := range branchResults {
			if b.op == dtmimp.OpCompensate && b.status != dtmcli.StatusSucceed &&
				branchResults[i+1].status != dtmcli.StatusPrepared {
				rsCToStart++
			}
		}
		logger.Debugf("rsCToStart: %d branchResults: %v", rsCToStart, branchResults)
	}
	timeLimit := time.Now().Add(time.Duration(conf.RequestTimeout+2) * time.Second)
	for time.Now().Before(timeLimit) && t.Status == dtmcli.StatusSubmitted && !t.isTimeout() && rsAFailed == 0 {
		toRun := pickToRunActions()
		runBranches(toRun)
		if rsADone == rsAStarted { // no branch is running, so break
			break
		}
		waitDoneOnce()
	}
	if t.Status == dtmcli.StatusSubmitted && rsAFailed == 0 && rsAToStart == rsASucceed {
		t.changeStatus(dtmcli.StatusSucceed)
		return nil
	}
	if t.Status == dtmcli.StatusSubmitted && rsAFailed > 0 {
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(failureError.Error()))
	}
	if t.Status == dtmcli.StatusSubmitted && t.isTimeout() {
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(fmt.Sprintf("Timeout after %d seconds", t.TimeoutToFail)))
	}
	if t.Status == dtmcli.StatusAborting {
		prepareToCompensate()
	}
	for time.Now().Before(timeLimit) && t.Status == dtmcli.StatusAborting {
		toRun := pickToRunCompensates()
		runBranches(toRun)
		if rsCDone == rsCToStart { // no branch is running, so break
			break
		}
		logger.Debugf("rsCDone: %d rsCToStart: %d", rsCDone, rsCToStart)
		waitDoneOnce()
	}
	if t.Status == dtmcli.StatusAborting && rsCToStart == rsCSucceed {
		t.changeStatus(dtmcli.StatusFailed)
	}
	return nil
}

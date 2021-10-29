package dtmsvr

import (
	"fmt"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

type transSagaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("saga", func(trans *TransGlobal) transProcessor { return &transSagaProcessor{TransGlobal: trans} })
}

func (t *transSagaProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	steps := []M{}
	dtmcli.MustUnmarshalString(t.Data, &steps)
	for i, step := range steps {
		branch := fmt.Sprintf("%02d", i+1)
		for _, branchType := range []string{dtmcli.BranchCompensate, dtmcli.BranchAction} {
			branches = append(branches, TransBranch{
				Gid:        t.Gid,
				BranchID:   branch,
				Data:       step["data"].(string),
				URL:        step[branchType].(string),
				BranchType: branchType,
				Status:     dtmcli.StatusPrepared,
			})
		}
	}
	return branches
}

type cSagaCustom struct {
	Orders     map[int][]int `json:"orders"`
	Concurrent bool          `json:"concurrent"`
}

type branchResult struct {
	index      int
	status     string
	started    bool
	branchType string
}

func (t *transSagaProcessor) ProcessOnce(db *common.DB, branches []TransBranch) error {
	// when saga tasks is fetched, it always need to process
	dtmcli.Logf("status: %s timeout: %t", t.Status, t.isTimeout())
	if t.Status == dtmcli.StatusSubmitted && t.isTimeout() {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	n := len(branches)

	csc := cSagaCustom{Orders: map[int][]int{}}
	if t.CustomData != "" {
		dtmcli.MustUnmarshalString(t.CustomData, &csc)
	}
	// resultStats
	var rsAToStart, rsAStarted, rsADone, rsAFailed, rsASucceed, rsCToStart, rsCDone, rsCSucceed int
	branchResults := make([]branchResult, n) // save the branch result
	for i := 0; i < n; i++ {
		b := branches[i]
		if b.BranchType == dtmcli.BranchAction {
			if b.Status == dtmcli.StatusPrepared {
				rsAToStart++
			} else if b.Status == dtmcli.StatusFailed {
				rsAFailed++
			}
		}
		branchResults[i] = branchResult{status: branches[i].Status, branchType: branches[i].BranchType}
	}
	isPreconditionsSucceed := func(current int) bool {
		// if !csc.Concurrent，then check the branch in previous step is succeed
		if !csc.Concurrent && current >= 2 && branches[current-2].Status != dtmcli.StatusSucceed {
			return false
		}
		// if csc.concurrent, then check the Orders. origin one step correspond to 2 step in dtmsvr
		for _, pre := range csc.Orders[current/2] {
			if branches[pre*2+1].Status != dtmcli.StatusSucceed {
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
				err = dtmcli.AsError(x)
			}
			resultChan <- branchResult{index: i, status: branches[i].Status, branchType: branches[i].BranchType}
			if err != nil {
				dtmcli.LogRedf("exec branch error: %v", err)
			}
		}()
		err = t.execBranch(db, &branches[i])
	}
	pickToRunActions := func() []int {
		toRun := []int{}
		for current := 0; current < n; current++ {
			br := &branchResults[current]
			if br.branchType == dtmcli.BranchAction && !br.started && isPreconditionsSucceed(current) && br.status == dtmcli.StatusPrepared {
				toRun = append(toRun, current)
			}
		}
		dtmcli.Logf("toRun picked for action is: %v", toRun)
		return toRun
	}
	runBranches := func(toRun []int) {
		for _, b := range toRun {
			branchResults[b].started = true
			if branchResults[b].branchType == dtmcli.BranchAction {
				rsAStarted++
			}
			go asyncExecBranch(b)
		}
	}
	pickAndRunCompensates := func(toRunActions []int) {
		for _, b := range toRunActions {
			// these branches may have run. so flag them to status succeed, then run the corresponding compensate
			branchResults[b].status = dtmcli.StatusSucceed
		}
		for i, b := range branchResults {
			if b.branchType == dtmcli.BranchCompensate && b.status != dtmcli.StatusSucceed && branchResults[i+1].status != dtmcli.StatusPrepared {
				rsCToStart++
				go asyncExecBranch(i)
			}
		}
	}
	waitDoneOnce := func() {
		select {
		case r := <-resultChan:
			br := &branchResults[r.index]
			br.status = r.status
			if r.branchType == dtmcli.BranchAction {
				rsADone++
				if r.status == dtmcli.StatusFailed {
					rsAFailed++
				} else if r.status == dtmcli.StatusSucceed {
					rsASucceed++
				}
			} else {
				rsCDone++
				if r.status == dtmcli.StatusSucceed {
					rsCSucceed++
				}
			}
			dtmcli.Logf("branch done: %v", r)
		case <-time.After(time.Duration(time.Second * 3)):
			dtmcli.Logf("wait once for done")
		}
	}

	for t.Status == dtmcli.StatusSubmitted && !t.isTimeout() && rsAFailed == 0 && rsADone != rsAToStart {
		toRun := pickToRunActions()
		runBranches(toRun)
		if rsADone == rsAStarted { // no branch is running, so break
			break
		}
		waitDoneOnce()
	}
	if t.Status == dtmcli.StatusSubmitted && rsAFailed == 0 && rsAToStart == rsASucceed {
		t.changeStatus(db, dtmcli.StatusSucceed)
		return nil
	}
	if t.Status == dtmcli.StatusSubmitted && (rsAFailed > 0 || t.isTimeout()) {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	if t.Status == dtmcli.StatusAborting {
		toRun := pickToRunActions()
		pickAndRunCompensates(toRun)
		for rsCDone != rsCToStart {
			waitDoneOnce()
		}
	}
	if t.Status == dtmcli.StatusAborting && rsCToStart == rsCSucceed {
		t.changeStatus(db, dtmcli.StatusFailed)
	}
	return nil
}

package dtmsvr

import (
	"fmt"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm/clause"
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
	if t.Status == dtmcli.StatusFailed || t.Status == dtmcli.StatusSucceed {
		return nil
	}
	if t.Status == dtmcli.StatusSubmitted && t.isTimeout() {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	n := len(branches)

	csc := cSagaCustom{Orders: map[int][]int{}}
	if t.CustomData != "" {
		dtmcli.MustUnmarshalString(t.CustomData, &csc)
	}
	// resultStats
	var rsActionToStart, rsActionDone, rsActionFailed, rsActionSucceed, rsCompensateToStart, rsCompensateDone, rsCompensateSucceed int
	branchResults := make([]branchResult, n) // save the branch result
	for i := 0; i < n; i++ {
		b := branches[i]
		if b.BranchType == dtmcli.BranchAction {
			if b.Status == dtmcli.StatusPrepared || b.Status == dtmcli.StatusDoing {
				rsActionToStart++
			} else if b.Status == dtmcli.StatusFailed {
				rsActionFailed++
			}
		}
		branchResults[i] = branchResult{status: branches[i].Status, branchType: branches[i].BranchType}
	}
	isPreconditionsSucceed := func(current int) bool {
		if !csc.Concurrent && branches[current-1].Status != dtmcli.StatusSucceed {
			return false
		}
		for _, pre := range csc.Orders[current*2+1] {
			if branches[pre*2+1].Status != dtmcli.StatusSucceed {
				return false
			}
		}
		return true
	}

	stopChan := make(chan branchResult, n)
	asyncExecBranch := func(i int) {
		var err error
		defer func() {
			if x := recover(); x != nil {
				err = dtmcli.AsError(x)
			}
			stopChan <- branchResult{index: i, status: branches[i].Status, branchType: branches[i].BranchType}
			if err != nil {
				dtmcli.LogRedf("exec branch error: %v", err)
			}
		}()
		err = t.execBranch(db, &branches[i])
	}
	needRollback := func(i int) bool {
		br := &branchResults[i]
		return !br.started && br.branchType == dtmcli.BranchCompensate && br.status != dtmcli.StatusSucceed && branchResults[i+1].branchType == dtmcli.BranchAction && branchResults[i+1].status != dtmcli.StatusPrepared
	}
	pickAndRun := func(branchType string) {
		toRun := []int{}
		for current := 0; current < n; current++ {
			br := &branchResults[current]
			if br.branchType == branchType && branchType == dtmcli.BranchAction {
				if (br.status == dtmcli.StatusPrepared || br.status == dtmcli.StatusDoing) &&
					!br.started && isPreconditionsSucceed(current) {
					br.status = dtmcli.StatusDoing
					toRun = append(toRun, current)
				}
			} else if br.branchType == branchType && branchType == dtmcli.BranchCompensate {
				if needRollback(current) {
					toRun = append(toRun, current)
				}
			}
		}
		if branchType == dtmcli.BranchAction && len(toRun) > 0 && csc.Concurrent { // only save doing when concurrent
			updates := make([]TransBranch, len(toRun))
			for i, b := range toRun {
				updates[i].ID = branches[b].ID
				branches[b].Status = dtmcli.StatusDoing
				updates[i].Status = dtmcli.StatusDoing
			}
			dbGet().Must().Clauses(clause.OnConflict{
				OnConstraint: "trans_branch_pkey",
				DoUpdates:    clause.AssignmentColumns([]string{"status"}),
			}).Create(updates)
		} else if branchType == dtmcli.BranchCompensate {
			// when not concurrent, then may add one more branch, in case the newest branch state not saved and timeout
			if !csc.Concurrent && len(toRun) < n/2 && branchResults[len(toRun)*2+1].status != dtmcli.StatusFailed {
				toRun = append(toRun, len(toRun)*2+2)
			}
			rsCompensateToStart = len(toRun)
		}
		for _, b := range toRun {
			branchResults[b].started = true
			go asyncExecBranch(b)
		}
	}
	processorTimeout := func() bool {
		return time.Since(t.processStarted)+NowForwardDuration > time.Duration(t.getRetryInterval()-3)*time.Second
	}
	waitOnceForDone := func() {
		select {
		case r := <-stopChan:
			br := &branchResults[r.index]
			br.status = r.status
			if r.branchType == dtmcli.BranchAction {
				rsActionDone++
				if r.status == dtmcli.StatusFailed {
					rsActionFailed++
				} else if r.status == dtmcli.StatusSucceed {
					rsActionSucceed++
				}
			} else {
				rsCompensateDone++
				if r.status == dtmcli.StatusSucceed {
					rsCompensateSucceed++
				}
			}
			dtmcli.Logf("branch done: %v", r)
		case <-time.After(time.Duration(time.Second * 3)):
			dtmcli.Logf("wait once for done")
		}
	}

	for t.Status == dtmcli.StatusSubmitted && !t.isTimeout() && rsActionFailed == 0 && rsActionDone != rsActionToStart && !processorTimeout() {
		pickAndRun(dtmcli.BranchAction)
		waitOnceForDone()
	}
	if t.Status == dtmcli.StatusSubmitted && rsActionFailed == 0 && rsActionToStart == rsActionSucceed {
		t.changeStatus(db, dtmcli.StatusSucceed)
		return nil
	}
	if t.Status == dtmcli.StatusSubmitted && (rsActionFailed > 0 || t.isTimeout()) {
		t.changeStatus(db, dtmcli.StatusAborting)
	}
	if t.Status == dtmcli.StatusAborting {
		pickAndRun(dtmcli.BranchCompensate)
		for rsCompensateDone != rsCompensateToStart && !processorTimeout() {
			waitOnceForDone()
		}
	}
	if (t.Status == dtmcli.StatusSubmitted || t.Status == dtmcli.StatusAborting) && rsActionFailed > 0 && rsCompensateToStart == rsCompensateSucceed {
		t.changeStatus(db, dtmcli.StatusFailed)
	}
	return nil
}

package dtmsvr

import (
	"context"
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
)

type transXaProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("xa", func(trans *TransGlobal) transProcessor { return &transXaProcessor{TransGlobal: trans} })
}

func (t *transXaProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *transXaProcessor) ProcessOnce(ctx context.Context, branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	if t.Status == dtmcli.StatusPrepared && t.isTimeout() {
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(fmt.Sprintf("Timeout after %d seconds", t.TimeoutToFail)))
	}
	currentType := dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmimp.OpCommit, dtmimp.OpRollback).(string)
	for i, branch := range branches {
		if branch.Op == currentType && branch.Status != dtmcli.StatusSucceed {
			err := t.execBranch(ctx, &branch, i)
			if err != nil {
				return err
			}
		}
	}
	t.changeStatus(dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

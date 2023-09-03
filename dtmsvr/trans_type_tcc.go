package dtmsvr

import (
	"context"
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/logger"
)

type transTccProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("tcc", func(trans *TransGlobal) transProcessor { return &transTccProcessor{TransGlobal: trans} })
}

func (t *transTccProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

func (t *transTccProcessor) ProcessOnce(ctx context.Context, branches []TransBranch) error {
	if !t.needProcess() {
		return nil
	}
	if t.Status == dtmcli.StatusPrepared && t.isTimeout() {
		t.changeStatus(dtmcli.StatusAborting, withRollbackReason(fmt.Sprintf("Timeout after %d seconds", t.TimeoutToFail)))
	}
	op := dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmimp.OpConfirm, dtmimp.OpCancel).(string)
	for current := len(branches) - 1; current >= 0; current-- {
		if branches[current].Op == op && branches[current].Status == dtmcli.StatusPrepared {
			logger.Debugf("branch info: current: %d ID: %d", current, branches[current].ID)
			err := t.execBranch(ctx, &branches[current], current)
			if err != nil {
				return err
			}
		}
	}
	t.changeStatus(dtmimp.If(t.Status == dtmcli.StatusSubmitted, dtmcli.StatusSucceed, dtmcli.StatusFailed).(string))
	return nil
}

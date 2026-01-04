package dtmsvr

import (
	"context"
	"math"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtm/client/workflow/wfpb"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
)

type transWorkflowProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("workflow", func(trans *TransGlobal) transProcessor { return &transWorkflowProcessor{TransGlobal: trans} })
}

func (t *transWorkflowProcessor) GenBranches() []TransBranch {
	return []TransBranch{}
}

type cWorkflowCustom struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

func (t *transWorkflowProcessor) ProcessOnce(ctx context.Context, branches []TransBranch) error {
	if t.Status == dtmcli.StatusFailed || t.Status == dtmcli.StatusSucceed {
		return nil
	}

	cmc := cWorkflowCustom{}
	dtmimp.MustUnmarshalString(t.CustomData, &cmc)
	data := cmc.Data
	if t.Protocol == dtmimp.ProtocolGRPC {
		wd := wfpb.WorkflowData{Data: cmc.Data}
		data = dtmgimp.MustProtoMarshal(&wd)
	}
	err := t.getURLResult(ctx, t.QueryPrepared, "00", cmc.Name, data)
	if err == dtmimp.ErrOngoing {
		t.touchCronTime(cronKeep, 0)
	} else if err != nil {
		t.touchCronTime(cronBackoff, 0)
		v := t.NextCronInterval / t.getNextCronInterval(cronReset)
		retryCount := int64(math.Log2(float64(v)))
		logger.Debugf("origin: %d v: %d retryCount: %d", t.getNextCronInterval(cronReset), v, retryCount)
		if retryCount >= conf.AlertRetryLimit && conf.AlertWebHook != "" {
			_, err2 := dtmcli.GetRestyClient().R().SetBody(gin.H{
				"gid":         t.Gid,
				"status":      t.Status,
				"branch":      "",
				"error":       err.Error(),
				"retry_count": retryCount,
			}).Post(conf.AlertWebHook)
			if err2 != nil {
				logger.Errorf("alerting webhook error: %v", err2)
			}
		}
	}
	return err
}

package dtmsvr

import (
	"context"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/logger"
	"github.com/gin-gonic/gin"
)

// TransGlobal global transaction
type TransGlobal struct {
	storage.TransGlobalStore
	ReqExtra         map[string]string `json:"req_extra"`
	Context          context.Context
	lastTouched      time.Time // record the start time of process
	updateBranchSync bool
}

func (t *TransGlobal) setupPayloads() {
	// Payloads will be store in BinPayloads, Payloads is only used to Unmarshal
	for _, p := range t.Payloads {
		t.BinPayloads = append(t.BinPayloads, []byte(p))
	}
	for _, d := range t.Steps {
		if d["data"] != "" {
			t.BinPayloads = append(t.BinPayloads, []byte(d["data"]))
		}
	}
	if t.Protocol == "" {
		t.Protocol = dtmimp.ProtocolHTTP
	}

}

// TransBranch branch transaction
type TransBranch = storage.TransBranchStore

type transProcessor interface {
	GenBranches() []TransBranch
	ProcessOnce(ctx context.Context, branches []TransBranch) error
}

type processorCreator func(*TransGlobal) transProcessor

var processorFac = map[string]processorCreator{}

func registorProcessorCreator(transType string, creator processorCreator) {
	processorFac[transType] = creator
}

func (t *TransGlobal) getProcessor() transProcessor {
	return processorFac[t.TransType](t)
}

type cronType int

const (
	cronBackoff cronType = iota
	cronReset
	cronKeep
)

// TransFromContext TransFromContext
func TransFromContext(c *gin.Context) *TransGlobal {
	b, err := c.GetRawData()
	e2p(err)
	m := TransGlobal{}
	dtmimp.MustUnmarshal(b, &m)
	m.Status = dtmimp.Escape(m.Status)
	m.Gid = dtmimp.Escape(m.Gid)
	logger.Debugf("creating trans in prepare")
	m.setupPayloads()
	m.Ext.Headers = map[string]string{}
	return &m
}

// TransFromDtmRequest TransFromContext
func TransFromDtmRequest(ctx context.Context, c *dtmgpb.DtmRequest) *TransGlobal {
	o := &dtmgpb.DtmTransOptions{}
	if c.TransOptions != nil {
		o = c.TransOptions
	}
	r := TransGlobal{TransGlobalStore: storage.TransGlobalStore{
		Gid:            c.Gid,
		TransType:      c.TransType,
		QueryPrepared:  c.QueryPrepared,
		Protocol:       "grpc",
		BinPayloads:    c.BinPayloads,
		CustomData:     c.CustomedData,
		RollbackReason: c.RollbackReason,
		TransOptions: dtmcli.TransOptions{
			WaitResult:     o.WaitResult,
			TimeoutToFail:  o.TimeoutToFail,
			RetryInterval:  o.RetryInterval,
			BranchHeaders:  o.BranchHeaders,
			RequestTimeout: o.RequestTimeout,
			RetryLimit:     o.RetryLimit,
		},
	}}
	r.ReqExtra = c.ReqExtra
	r.Context = ctx
	if c.Steps != "" {
		dtmimp.MustUnmarshalString(c.Steps, &r.Steps)
	}
	return &r
}

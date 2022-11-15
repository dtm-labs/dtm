package dtmgrpc

import (
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
)

// TransBaseOption setup func for TransBase
type TransBaseOption func(tb *dtmimp.TransBase)

// WithBranchHeaders setup TransBase.BranchHeaders
func WithBranchHeaders(headers map[string]string) TransBaseOption {
	return func(tb *dtmimp.TransBase) {
		tb.BranchHeaders = headers
	}
}

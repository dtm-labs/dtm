package workflow

import (
	"context"
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/stretchr/testify/assert"
)

func TestAbnormal(t *testing.T) {
	fname := dtmimp.GetFuncName()
	_, err := defaultFac.execute(context.Background(), fname, fname, nil)
	assert.Error(t, err)

	err = defaultFac.register(fname, func(wf *Workflow, data []byte) ([]byte, error) { return nil, nil })
	assert.Nil(t, err)
	err = defaultFac.register(fname, nil)
	assert.Error(t, err)

	ws := &workflowServer{}
	_, err = ws.Execute(context.Background(), nil)
	assert.Contains(t, err.Error(), "call workflow.InitGrpc first")
}

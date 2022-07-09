package workflow

import (
	"context"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/stretchr/testify/assert"
)

func TestAbnormal(t *testing.T) {
	fname := dtmimp.GetFuncName()
	err := defaultFac.execute(fname, fname, nil)
	assert.Error(t, err)

	err = defaultFac.register(fname, func(wf *Workflow, data []byte) error { return nil })
	assert.Nil(t, err)
	err = defaultFac.register(fname, nil)
	assert.Error(t, err)

	ws := &workflowServer{}
	_, err = ws.Execute(context.Background(), nil)
	assert.Contains(t, err.Error(), "call workflow.InitGrpc first")
}

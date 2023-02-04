package test

import (
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/workflow"
	"github.com/dtm-labs/dtm/test/busi"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowRet(t *testing.T) {
	workflow.SetProtocolForTest(dtmimp.ProtocolHTTP)
	req := busi.GenReqHTTP(30, false, false)
	gid := dtmimp.GetFuncName()

	workflow.Register2(gid, func(wf *workflow.Workflow, data []byte) ([]byte, error) {
		var req busi.ReqHTTP
		dtmimp.MustUnmarshal(data, &req)
		_, err := wf.NewBranch().NewRequest().SetBody(req).Post(Busi + "/TransOut")
		return []byte("result of workflow"), err
	})

	ret, err := workflow.Execute2(gid, gid, dtmimp.MustMarshal(req))
	assert.Nil(t, err)
	assert.Equal(t, "result of workflow", string(ret))
	assert.Equal(t, StatusSucceed, getTransStatus(gid))

	// the second execute will return result directly
	ret, err = workflow.Execute2(gid, gid, dtmimp.MustMarshal(req))
	assert.Nil(t, err)
	assert.Equal(t, "result of workflow", string(ret))
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}

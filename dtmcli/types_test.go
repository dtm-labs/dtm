package dtmcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
)

func TestTypes(t *testing.T) {
	err := common.CatchP(func() {
		idGen := IDGenerator{parentID: "12345678901234567890123"}
		idGen.NewBranchID()
	})
	assert.Error(t, err)
	err = common.CatchP(func() {
		idGen := IDGenerator{branchID: 99}
		idGen.NewBranchID()
	})
	assert.Error(t, err)
}

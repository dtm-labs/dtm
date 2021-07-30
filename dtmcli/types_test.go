package dtmcli

import (
	"fmt"
	"net/url"
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
	_, err = TransInfoFromQuery(url.Values{})
	assert.Error(t, err)

	err2 := fmt.Errorf("an error")
	err3 := CheckDtmResponse(nil, err2)
	assert.Error(t, err2, err3)
}

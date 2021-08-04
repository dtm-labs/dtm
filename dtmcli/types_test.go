package dtmcli

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	err := CatchP(func() {
		idGen := IDGenerator{parentID: "12345678901234567890123"}
		idGen.NewBranchID()
	})
	assert.Error(t, err)
	err = CatchP(func() {
		idGen := IDGenerator{branchID: 99}
		idGen.NewBranchID()
	})
	err = CatchP(func() {
		MustGenGid("http://localhost:8080/api/no")
	})
	assert.Error(t, err)
	assert.Error(t, err)
	_, err = TransInfoFromQuery(url.Values{})
	assert.Error(t, err)

}

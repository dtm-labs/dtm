package dtmcli

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := common.MS{}
	resp, err := common.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// CheckDtmResponse check the response of dtm, if not ok ,generate error
func CheckDtmResponse(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}
	if !strings.Contains(resp.String(), "SUCCESS") {
		return fmt.Errorf("dtm response failed: %s", resp.String())
	}
	return nil
}

// IDGenerator used to generate a branch id
type IDGenerator struct {
	parentID string
	branchID int
}

// NewBranchID generate a branch id
func (g *IDGenerator) NewBranchID() string {
	if g.branchID >= 99 {
		panic(fmt.Errorf("branch id is larger than 99"))
	}
	if len(g.parentID) >= 20 {
		panic(fmt.Errorf("total branch id is longer than 20"))
	}
	g.branchID = g.branchID + 1
	return g.parentID + fmt.Sprintf("%02d", g.branchID)
}

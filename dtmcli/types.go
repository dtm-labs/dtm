package dtmcli

import (
	"fmt"

	"github.com/yedf/dtm/common"
)

func GenGid(server string) string {
	res := common.MS{}
	_, err := common.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	e2p(err)
	return res["gid"]
}

type IDGenerator struct {
	parentID string
	branchID int
}

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

package dtmcli

import "github.com/yedf/dtm/common"

func GenGid(server string) string {
	res := common.MS{}
	_, err := common.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	e2p(err)
	return res["gid"]
}

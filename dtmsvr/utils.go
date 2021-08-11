package dtmsvr

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// M a short name
type M = map[string]interface{}

var p2e = dtmcli.P2E
var e2p = dtmcli.E2P

var config = common.DtmConfig

func dbGet() *common.DB {
	return common.DbGet(config.DB)
}
func writeTransLog(gid string, action string, status string, branch string, detail string) {
	// if detail == "" {
	// 	detail = "{}"
	// }
	// dbGet().Must().Table("trans_log").Create(M{
	// 	"gid":        gid,
	// 	"action":     action,
	// 	"new_status": status,
	// 	"branch":     branch,
	// 	"detail":     detail,
	// })
}

// TransProcessedTestChan only for test usage. when transaction processed once, write gid to this chan
var TransProcessedTestChan chan string = nil

// WaitTransProcessed only for test usage. wait for transaction processed once
func WaitTransProcessed(gid string) {
	dtmcli.Logf("waiting for gid %s", gid)
	select {
	case id := <-TransProcessedTestChan:
		for id != gid {
			dtmcli.LogRedf("-------id %s not match gid %s", id, gid)
			id = <-TransProcessedTestChan
		}
		dtmcli.Logf("finish for gid %s", gid)
	case <-time.After(time.Duration(time.Second * 3)):
		dtmcli.LogFatalf("Wait Trans timeout")
	}
}

var gNode *snowflake.Node = nil

func init() {
	node, err := snowflake.NewNode(1)
	e2p(err)
	gNode = node
}

// GenGid generate gid, use ip + snowflake
func GenGid() string {
	return getOneHexIP() + "_" + gNode.Generate().Base58()
}

func getOneHexIP() string {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4().String()
				ns := strings.Split(ip, ".")
				r := []byte{}
				for _, n := range ns {
					r = append(r, byte(dtmcli.MustAtoi(n)))
				}
				return hex.EncodeToString(r)
			}
		}
	}
	fmt.Printf("err is: %s", err.Error())
	return "" // 获取不到IP，则直接返回空
}

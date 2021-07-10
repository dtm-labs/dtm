package dtmsvr

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

type M = map[string]interface{}

var p2e = common.P2E
var e2p = common.E2P

func dbGet() *common.DB {
	return common.DbGet(config.Mysql)
}
func writeTransLog(gid string, action string, status string, branch string, detail string) {
	return
	db := dbGet()
	if detail == "" {
		detail = "{}"
	}
	db.Must().Table("trans_log").Create(M{
		"gid":        gid,
		"action":     action,
		"new_status": status,
		"branch":     branch,
		"detail":     detail,
	})
}

var TransProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitTransProcessed(gid string) {
	logrus.Printf("waiting for gid %s", gid)
	id := <-TransProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-TransProcessedTestChan
	}
	logrus.Printf("finish for gid %s", gid)
}

var gNode *snowflake.Node = nil

func init() {
	node, err := snowflake.NewNode(1)
	e2p(err)
	gNode = node
}

func GenGid() string {
	return getOneHexIp() + "-" + gNode.Generate().Base58()
}

func getOneHexIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("cannot get ip, default to another call")
		return gNode.Generate().Base58()
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4().String()
				ns := strings.Split(ip, ".")
				r := []byte{}
				for _, n := range ns {
					r = append(r, byte(common.MustAtoi(n)))
				}
				return hex.EncodeToString(r)
			}

		}
	}
	fmt.Printf("none ipv4, default to another call")
	return gNode.Generate().Base58()
}

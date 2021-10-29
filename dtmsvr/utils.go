package dtmsvr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm"
)

// M a short name
type M = map[string]interface{}

type branchStatus struct {
	id         uint64
	status     string
	finishTime *time.Time
}

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
	// 	dtmcli.BranchAction:     action,
	// 	"new_status": status,
	// 	"branch":     branch,
	// 	"detail":     detail,
	// })
}

// TransProcessedTestChan only for test usage. when transaction processed once, write gid to this chan
var TransProcessedTestChan chan string = nil

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

// transFromDb construct trans from db
func transFromDb(db *common.DB, gid string) *TransGlobal {
	m := TransGlobal{}
	dbr := db.Must().Model(&m).Where("gid=?", gid).First(&m)
	if dbr.Error == gorm.ErrRecordNotFound {
		return nil
	}
	e2p(dbr.Error)
	return &m
}

func checkLocalhost(branches []TransBranch) {
	if config.DisableLocalhost == 0 {
		return
	}
	for _, branch := range branches {
		if strings.HasPrefix(branch.URL, "http://localhost") || strings.HasPrefix(branch.URL, "localhost") {
			panic(errors.New("url for localhost is disabled. check for your config"))
		}
	}
}

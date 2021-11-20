/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

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
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type branchStatus struct {
	id         uint64
	status     string
	finishTime *time.Time
}

var p2e = dtmimp.P2E
var e2p = dtmimp.E2P

var config = common.DtmConfig

func dbGet() *common.DB {
	return common.DbGet(config.DB)
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
					r = append(r, byte(dtmimp.MustAtoi(n)))
				}
				return hex.EncodeToString(r)
			}
		}
	}
	fmt.Printf("err is: %s", err.Error())
	return "" // 获取不到IP，则直接返回空
}

// transFromDb construct trans from db
func transFromDb(db *gorm.DB, gid string, lock bool) *TransGlobal {
	m := TransGlobal{}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	dbr := db.Model(&m).Where("gid=?", gid).First(&m)
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

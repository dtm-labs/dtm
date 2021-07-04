package dtmsvr

import (
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

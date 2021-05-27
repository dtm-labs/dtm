package dtmsvr

import (
	"github.com/sirupsen/logrus"
)

var TransProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitTransProcessed(gid string) {
	id := <-TransProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-TransProcessedTestChan
	}
}

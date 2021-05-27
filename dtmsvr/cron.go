package dtmsvr

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func CronPreparedOnce(expire time.Duration) {
	db := dbGet()
	ss := []TransGlobal{}
	db.Must().Model(&TransGlobal{}).Where("update_time < date_sub(now(), interval ? second)", int(expire/time.Second)).Where("status = ?", "prepared").Find(&ss)
	writeTransLog("", "saga fetch prepared", fmt.Sprint(len(ss)), "", "")
	if len(ss) == 0 {
		return
	}
	for _, sm := range ss {
		writeTransLog(sm.Gid, "saga touch prepared", "", "", "")
		db.Must().Model(&sm).Update("id", sm.ID)
		resp, err := common.RestyClient.R().SetQueryParam("gid", sm.Gid).Get(sm.QueryPrepared)
		e2p(err)
		body := resp.String()
		if strings.Contains(body, "FAIL") {
			preparedExpire := time.Now().Add(time.Duration(-config.PreparedExpire) * time.Second)
			logrus.Printf("create time: %s prepared expire: %s ", sm.CreateTime.Local(), preparedExpire.Local())
			status := common.If(sm.CreateTime.Before(preparedExpire), "canceled", "prepared").(string)
			writeTransLog(sm.Gid, "saga canceled", status, "", "")
			db.Must().Model(&sm).Where("status = ?", "prepared").Update("status", status)
		} else if strings.Contains(body, "SUCCESS") {
			saveCommitted(&sm)
			ProcessTrans(&sm)
		}
	}
}

func CronPrepared() {
	for {
		defer handlePanic()
		CronTransOnce(time.Duration(config.JobCronInterval)*time.Second, "prepared")
		sleepCronTime()
	}
}

func CronTransOnce(expire time.Duration, status string) bool {
	trans := lockOneTrans(expire, status)
	if trans == nil {
		return false
	}
	trans.touch(dbGet())
	branches := []TransBranch{}
	db := dbGet()
	db.Must().Where("gid=?", trans.Gid).Order("id asc").Find(&branches)
	trans.getProcessor().ProcessOnce(db, branches)
	if TransProcessedTestChan != nil {
		TransProcessedTestChan <- trans.Gid
	}
	return true
}

func CronCommitted() {
	for {
		defer handlePanic()
		processed := CronTransOnce(time.Duration(config.JobCronInterval)*time.Second, "commitetd")
		if !processed {
			sleepCronTime()
		}
	}
}

func lockOneTrans(expire time.Duration, status string) *TransGlobal {
	trans := TransGlobal{}
	owner := common.GenGid()
	db := dbGet()
	dbr := db.Must().Model(&trans).
		Where("update_time < date_sub(now(), interval ? second) and satus=?", int(expire/time.Second), status).
		Limit(1).Update("owner", owner)
	if dbr.RowsAffected == 0 {
		return nil
	}
	dbr = db.Must().Where("owner=?", owner).Find(&trans)
	return &trans
}

func handlePanic() {
	if err := recover(); err != nil {
		logrus.Printf("----panic %s handlered", err.(error).Error())
		time.Sleep(3 * time.Second) // 出错后睡眠3s，避免无限循环
	}
}

func sleepCronTime() {
	delta := math.Min(3, float64(config.JobCronInterval))
	interval := time.Duration(rand.Float64() * delta * float64(time.Second))
	time.Sleep(interval)
}

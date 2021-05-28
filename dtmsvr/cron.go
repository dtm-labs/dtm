package dtmsvr

import (
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func CronPrepared() {
	for {
		CronTransOnce(time.Duration(config.JobCronInterval)*time.Second, "prepared")
		sleepCronTime()
	}
}

func CronTransOnce(expire time.Duration, status string) bool {
	defer handlePanic()
	trans := lockOneTrans(expire, status)
	if trans == nil {
		return false
	}
	trans.touch(dbGet())
	defer func() {
		WaitTransProcessed(trans.Gid)
	}()
	trans.Process(dbGet())
	return true
}

func CronCommitted() {
	for {
		notEmpty := CronTransOnce(time.Duration(config.JobCronInterval)*time.Second, "commitetd")
		if !notEmpty {
			sleepCronTime()
		}
	}
}

func lockOneTrans(expire time.Duration, status string) *TransGlobal {
	trans := TransGlobal{}
	owner := common.GenGid()
	db := dbGet()
	dbr := db.Must().Model(&trans).
		Where("update_time < date_sub(now(), interval ? second) and status=?", int(expire/time.Second), status).
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
	}
}

func sleepCronTime() {
	delta := math.Min(3, float64(config.JobCronInterval))
	interval := time.Duration(rand.Float64() * delta * float64(time.Second))
	time.Sleep(interval)
}

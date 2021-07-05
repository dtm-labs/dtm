package dtmsvr

import (
	"math"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func CronPrepared() {
	for {
		CronTransOnce(time.Duration(0), "prepared")
		sleepCronTime()
	}
}

func CronTransOnce(expireIn time.Duration, status string) bool {
	defer handlePanic()
	trans := lockOneTrans(expireIn, status)
	if trans == nil {
		return false
	}
	defer WaitTransProcessed(trans.Gid)
	trans.Process(dbGet())
	return true
}

func CronCommitted() {
	for {
		notEmpty := CronTransOnce(time.Duration(0), "commitetd")
		if !notEmpty {
			sleepCronTime()
		}
	}
}

func lockOneTrans(expireIn time.Duration, status string) *TransGlobal {
	trans := TransGlobal{}
	owner := common.GenGid()
	db := dbGet()
	dbr := db.Must().Model(&trans).
		Where("next_cron_time < date_add(now(), interval ? second) and status=?", int(expireIn/time.Second), status).
		Limit(1).Update("owner", owner)
	if dbr.RowsAffected == 0 {
		return nil
	}
	dbr = db.Must().Where("owner=?", owner).Find(&trans)
	updates := trans.setNextCron(trans.NextCronInterval * 2) // 下次被cron的间隔加倍
	db.Must().Model(&trans).Select(updates).Updates(&trans)
	return &trans
}

func handlePanic() {
	if err := recover(); err != nil {
		logrus.Errorf("\x1b[31m\n----panic %s handlered\x1b[0m\n%s", err.(error).Error(), string(debug.Stack()))
	}
}

func sleepCronTime() {
	delta := math.Min(3, float64(config.TransCronInterval))
	interval := time.Duration((float64(config.TransCronInterval) - rand.Float64()*delta) * float64(time.Second))
	time.Sleep(interval)
}

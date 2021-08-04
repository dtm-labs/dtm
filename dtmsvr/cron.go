package dtmsvr

import (
	"fmt"
	"math"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/yedf/dtm/dtmcli"
)

// CronTransOnce cron expired trans. use expireIn as expire time
func CronTransOnce(expireIn time.Duration) bool {
	defer handlePanic(nil)
	trans := lockOneTrans(expireIn)
	if trans == nil {
		return false
	}
	if TransProcessedTestChan != nil {
		defer WaitTransProcessed(trans.Gid)
	}
	trans.Process(dbGet(), true)
	return true
}

// CronExpiredTrans cron expired trans, num == -1 indicate for ever
func CronExpiredTrans(num int) {
	for i := 0; i < num || num == -1; i++ {
		hasTrans := CronTransOnce(time.Duration(0))
		if !hasTrans && num != 1 {
			sleepCronTime()
		}
	}
}

func lockOneTrans(expireIn time.Duration) *TransGlobal {
	trans := TransGlobal{}
	owner := GenGid()
	db := dbGet()
	// 这里next_cron_time需要限定范围，否则数据量累计之后，会导致查询变慢
	// 限定update_time < now - 3，否则会出现刚被这个应用取出，又被另一个取出
	dbr := db.Must().Model(&trans).
		Where("next_cron_time < date_add(now(), interval ? second) and next_cron_time > date_add(now(), interval -3600 second) and update_time < date_add(now(), interval ? second) and status in ('prepared', 'aborting', 'submitted')", int(expireIn/time.Second), -3+int(expireIn/time.Second)).
		Limit(1).Update("owner", owner)
	if dbr.RowsAffected == 0 {
		return nil
	}
	dbr = db.Must().Where("owner=?", owner).Find(&trans)
	updates := trans.setNextCron(trans.NextCronInterval * 2) // 下次被cron的间隔加倍
	db.Must().Model(&trans).Select(updates).Updates(&trans)
	return &trans
}

func handlePanic(perr *error) {
	if err := recover(); err != nil {
		dtmcli.LogRedf("----panic %v handlered\n%s", err, string(debug.Stack()))
		if perr != nil {
			*perr = fmt.Errorf("dtm panic: %v", err)
		}
	}
}

func sleepCronTime() {
	delta := math.Min(3, float64(config.TransCronInterval))
	interval := time.Duration((float64(config.TransCronInterval) - rand.Float64()*delta) * float64(time.Second))
	dtmcli.Logf("sleeping for %v", interval)
	time.Sleep(interval)
}

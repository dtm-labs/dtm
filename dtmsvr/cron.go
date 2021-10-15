package dtmsvr

import (
	"fmt"
	"math"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/yedf/dtm/dtmcli"
)

// CronForwardDuration will be set in test, cron will fetch trans which expire in CronForwardDuration
var CronForwardDuration time.Duration = time.Duration(0)

// CronTransOnce cron expired trans. use expireIn as expire time
func CronTransOnce() (hasTrans bool) {
	defer handlePanic(nil)
	trans := lockOneTrans(CronForwardDuration)
	if trans == nil {
		return
	}
	hasTrans = true
	if TransProcessedTestChan != nil {
		defer WaitTransProcessed(trans.Gid)
	}
	trans.Process(dbGet(), true)
	return
}

// CronExpiredTrans cron expired trans, num == -1 indicate for ever
func CronExpiredTrans(num int) {
	for i := 0; i < num || num == -1; i++ {
		hasTrans := CronTransOnce()
		if !hasTrans && num != 1 {
			sleepCronTime(0)
		}
	}
}

func lockOneTrans(expireIn time.Duration) *TransGlobal {
	trans := TransGlobal{}
	owner := GenGid()
	db := dbGet()
	getTime := dtmcli.GetDBSpecial().TimestampAdd
	expire := int(expireIn / time.Second)
	whereTime := fmt.Sprintf("next_cron_time < %s and next_cron_time > %s and update_time < %s", getTime(expire), getTime(-3600), getTime(expire-3))
	// 这里next_cron_time需要限定范围，否则数据量累计之后，会导致查询变慢
	// 限定update_time < now - 3，否则会出现刚被这个应用取出，又被另一个取出
	dbr := db.Must().Model(&trans).
		Where(whereTime+"and status in ('prepared', 'aborting', 'submitted')").Limit(1).Update("owner", owner)
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
		dtmcli.LogRedf("----recovered panic %v\n%s", err, string(debug.Stack()))
		if perr != nil {
			*perr = fmt.Errorf("dtm panic: %v", err)
		}
	}
}

func sleepCronTime(milli int) {
	delta := math.Min(3, float64(config.TransCronInterval))
	interval := time.Duration((float64(config.TransCronInterval) - rand.Float64()*delta) * float64(time.Second))
	dtmcli.Logf("sleeping for %v pass in %d milli", interval, milli)
	time.Sleep(dtmcli.If(milli == 0, interval, time.Duration(milli*int(time.Millisecond))).(time.Duration))
}

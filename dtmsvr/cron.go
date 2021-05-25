package dtmsvr

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func CronPreparedOnce(expire time.Duration) {
	db := dbGet()
	ss := []TransGlobalModel{}
	db.Must().Model(&TransGlobalModel{}).Where("update_time < date_sub(now(), interval ? second)", int(expire/time.Second)).Where("status = ?", "prepared").Find(&ss)
	writeTransLog("", "saga fetch prepared", fmt.Sprint(len(ss)), "", "")
	if len(ss) == 0 {
		return
	}
	for _, sm := range ss {
		writeTransLog(sm.Gid, "saga touch prepared", "", "", "")
		db.Must().Model(&sm).Update("id", sm.ID)
		resp, err := common.RestyClient.R().SetQueryParam("gid", sm.Gid).Get(sm.QueryPrepared)
		common.PanicIfError(err)
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
		CronPreparedOnce(10 * time.Second)
	}
}

func CronCommittedOnce(expire time.Duration) {
	db := dbGet()
	ss := []TransGlobalModel{}
	db.Must().Model(&TransGlobalModel{}).Where("update_time < date_sub(now(), interval ? second)", int(expire/time.Second)).Where("status = ?", "committed").Find(&ss)
	writeTransLog("", "saga fetch committed", fmt.Sprint(len(ss)), "", "")
	if len(ss) == 0 {
		return
	}
	for _, sm := range ss {
		writeTransLog(sm.Gid, "saga touch committed", "", "", "")
		db.Must().Model(&sm).Update("id", sm.ID)
		ProcessTrans(&sm)
	}
}

func CronCommitted() {
	for {
		defer handlePanic()
		CronCommittedOnce(10 * time.Second)
	}
}

func handlePanic() {
	if err := recover(); err != nil {
		logrus.Printf("----panic %s handlered", err.(error).Error())
	}
}

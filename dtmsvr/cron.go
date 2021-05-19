package dtmsvr

import (
	"strings"
	"time"

	"github.com/yedf/dtm/common"
)

func CronPreparedOne(expire time.Duration) {
	db := DbGet()
	sm := SagaModel{}
	dbr := db.Model(&sm).Where("update_time > date_add(now(), interval ? second)", int(expire/time.Second)).Where("status = ?", "prepared").First(&sm)
	common.PanicIfError(dbr.Error)
	resp, err := common.RestyClient.R().SetQueryParam("gid", sm.Gid).Get(sm.TransQuery)
	common.PanicIfError(err)
	body := resp.String()
	if strings.Contains(body, "FAIL") {
		dbr = db.Model(&sm).Where("status = ?", "prepared").Update("status", "canceled")
		common.PanicIfError(dbr.Error)
	} else if strings.Contains(body, "SUCESS") {
		dbr = db.Model(&sm).Where("status = ?", "")
	}
}

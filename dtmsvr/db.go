package dtmsvr

import "github.com/yedf/dtm/common"

func dbGet() *common.MyDb {
	return common.DbGet(config.Mysql)
}
func writeTransLog(gid string, action string, status string, branch string, detail string) {
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

package dtmsvr

import "fmt"

func svcSubmit(t *TransGlobal, waitResult bool) (interface{}, error) {
	db := dbGet()
	dbt := TransFromDb(db, t.Gid)
	if dbt != nil && dbt.Status != "prepared" && dbt.Status != "submitted" {
		return M{"dtm_result": "FAILURE", "message": fmt.Sprintf("current status %s, cannot sumbmit", dbt.Status)}, nil
	}
	t.Status = "submitted"
	t.saveNew(db)
	return t.Process(db, waitResult), nil

}

package dtmsvr

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func saveCommitted(m *TransGlobalModel) {
	db := dbGet()
	m.Status = "committed"
	err := db.Transaction(func(db1 *gorm.DB) error {
		db := &common.MyDb{DB: db1}
		writeTransLog(m.Gid, "save committed", m.Status, "", m.Data)
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(m)
		if dbr.RowsAffected == 0 {
			writeTransLog(m.Gid, "change status", m.Status, "", "")
			db.Must().Model(m).Where("status=?", "prepared").Update("status", "committed")
		}
		nsteps := GetTrans(m).GetDataBranches()
		if len(nsteps) > 0 {
			writeTransLog(m.Gid, "save steps", m.Status, "", common.MustMarshalString(nsteps))
			db.Must().Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&nsteps)
		}
		return nil
	})
	common.PanicIfError(err)
}

var TransProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitTransProcessed(gid string) {
	id := <-TransProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-TransProcessedTestChan
	}
}

func ProcessTrans(trans *TransGlobalModel) {
	err := innerProcessTrans(trans)
	if err != nil {
		logrus.Errorf("process trans ignore error: %s", err.Error())
	}
	if TransProcessedTestChan != nil {
		TransProcessedTestChan <- trans.Gid
	}
}
func innerProcessTrans(trans *TransGlobalModel) (rerr error) {
	branches := []TransBranchModel{}
	db := dbGet()
	db.Must().Where("gid=?", trans.Gid).Order("id asc").Find(&branches)
	return GetTrans(trans).ProcessOnce(db, branches)
}

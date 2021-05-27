package dtmsvr

import (
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func saveCommitted(m *TransGlobal) {
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
		nsteps := m.getProcessor().GenBranches()
		if len(nsteps) > 0 {
			writeTransLog(m.Gid, "save steps", m.Status, "", common.MustMarshalString(nsteps))
			db.Must().Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&nsteps)
		}
		return nil
	})
	e2p(err)
}

var TransProcessedTestChan chan string = nil // 用于测试时，通知处理结束

func WaitTransProcessed(gid string) {
	id := <-TransProcessedTestChan
	for id != gid {
		logrus.Errorf("-------id %s not match gid %s", id, gid)
		id = <-TransProcessedTestChan
	}
}

func ProcessTrans(trans *TransGlobal) {
	branches := []TransBranch{}
	db := dbGet()
	db.Must().Where("gid=?", trans.Gid).Order("id asc").Find(&branches)
	trans.getProcessor().ProcessOnce(db, branches)
	if TransProcessedTestChan != nil {
		TransProcessedTestChan <- trans.Gid
	}
}

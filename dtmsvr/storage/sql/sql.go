package sql

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
)

var config = &common.Config

type SqlStore struct {
}

func (s *SqlStore) Ping() error {
	db, err := dtmimp.StandaloneDB(config.Store.GetDBConf())
	dtmimp.E2P(err)
	_, err = db.Exec("select 1")
	return err
}

func (s *SqlStore) PopulateData(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.storage.%s.sql", common.GetSqlDir(), config.Store.Driver)
	common.RunSQLScript(config.Store.GetDBConf(), file, skipDrop)
}

func (s *SqlStore) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	trans := &storage.TransGlobalStore{}
	dbr := dbGet().Model(trans).Where("gid=?", gid).First(trans)
	if dbr.Error == gorm.ErrRecordNotFound {
		return nil
	}
	dtmimp.E2P(dbr.Error)
	return trans
}

func (s *SqlStore) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
	globals := []storage.TransGlobalStore{}
	lid := math.MaxInt64
	if *position != "" {
		lid = dtmimp.MustAtoi(*position)
	}
	dbr := dbGet().Must().Where("id < ?", lid).Order("id desc").Limit(int(limit)).Find(&globals)
	if dbr.RowsAffected < limit {
		*position = ""
	} else {
		*position = fmt.Sprintf("%d", globals[len(globals)-1].ID)
	}
	return globals
}

func (s *SqlStore) FindBranches(gid string) []storage.TransBranchStore {
	branches := []storage.TransBranchStore{}
	dbGet().Must().Where("gid=?", gid).Order("id asc").Find(&branches)
	return branches
}

func (s *SqlStore) UpdateBranchesSql(branches []storage.TransBranchStore, updates []string) *gorm.DB {
	return dbGet().Clauses(clause.OnConflict{
		OnConstraint: "trans_branch_op_pkey",
		DoUpdates:    clause.AssignmentColumns(updates),
	}).Create(branches)
}

func (s *SqlStore) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {
	err := dbGet().Transaction(func(tx *gorm.DB) error {
		g := &storage.TransGlobalStore{}
		dbr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(g).Where("gid=? and status=?", gid, status).First(g)
		if dbr.Error == nil {
			dbr = tx.Save(branches)
		}
		return wrapError(dbr.Error)
	})
	dtmimp.E2P(err)
}

func (s *SqlStore) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	return dbGet().Transaction(func(db1 *gorm.DB) error {
		db := &common.DB{DB: db1}
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(global)
		if dbr.RowsAffected <= 0 { // 如果这个不是新事务，返回错误
			return storage.ErrUniqueConflict
		}
		if len(branches) > 0 {
			db.Must().Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&branches)
		}
		return nil
	})
}

func (s *SqlStore) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	dbr := dbGet().Must().Model(global).Where("status=? and gid=?", old, global.Gid).Select(updates).Updates(global)
	if dbr.RowsAffected == 0 {
		dtmimp.E2P(storage.ErrNotFound)
	}
}

func (s *SqlStore) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64) {
	global.NextCronTime = common.GetNextTime(nextCronInterval)
	global.UpdateTime = common.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	dbGet().Must().Model(global).Where("status=? and gid=?", global.Status, global.Gid).
		Select([]string{"next_cron_time", "update_time", "next_cron_interval"}).Updates(global)
}

func (s *SqlStore) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	db := dbGet()
	getTime := func(second int) string {
		return map[string]string{
			"mysql":    fmt.Sprintf("date_add(now(), interval %d second)", second),
			"postgres": fmt.Sprintf("current_timestamp + interval '%d second'", second),
		}[config.Store.Driver]
	}
	expire := int(expireIn / time.Second)
	whereTime := fmt.Sprintf("next_cron_time < %s", getTime(expire))
	owner := uuid.NewString()
	global := &storage.TransGlobalStore{}
	dbr := db.Must().Model(global).
		Where(whereTime + "and status in ('prepared', 'aborting', 'submitted')").
		Limit(1).
		Select([]string{"owner", "next_cron_time"}).
		Updates(&storage.TransGlobalStore{
			Owner:        owner,
			NextCronTime: common.GetNextTime(common.Config.RetryInterval),
		})
	if dbr.RowsAffected == 0 {
		return nil
	}
	dbr = db.Must().Where("owner=?", owner).First(global)
	return global
}

func dbGet() *common.DB {
	return common.DbGet(config.Store.GetDBConf())
}

func wrapError(err error) error {
	if err == gorm.ErrRecordNotFound {
		return storage.ErrNotFound
	}
	dtmimp.E2P(err)
	return err
}

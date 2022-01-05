package sql

import (
	"fmt"
	"math"
	"time"

	"github.com/lithammer/shortuuid/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
)

var conf = &config.Config

// Store implements storage.Store feature that for DB
type Store struct{}

// Ping execs ping cmd to db
func (s *Store) Ping() error {
	db, err := dtmimp.StandaloneDB(conf.Store.GetDBConf())
	dtmimp.E2P(err)
	_, err = db.Exec("select 1")
	return err
}

// PopulateData populates data to db
func (s *Store) PopulateData(skipDrop bool) {
	file := fmt.Sprintf("%s/dtmsvr.storage.%s.sql", dtmutil.GetSQLDir(), conf.Store.Driver)
	dtmutil.RunSQLScript(conf.Store.GetDBConf(), file, skipDrop)
}

// FindTransGlobalStore finds TransGlobalStore by gid
func (s *Store) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	trans := &storage.TransGlobalStore{}
	dbr := dbGet().Model(trans).Where("gid=?", gid).First(trans)
	if dbr.Error == gorm.ErrRecordNotFound {
		return nil
	}
	dtmimp.E2P(dbr.Error)
	return trans
}

// ScanTransGlobalStores lists TransGlobalStore
func (s *Store) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
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

// FindBranches finds TransBranchStore by gid
func (s *Store) FindBranches(gid string) []storage.TransBranchStore {
	branches := []storage.TransBranchStore{}
	dbGet().Must().Where("gid=?", gid).Order("id asc").Find(&branches)
	return branches
}

// UpdateBranches updates TransBranchStore
func (s *Store) UpdateBranches(branches []storage.TransBranchStore, updates []string) (int, error) {
	db := dbGet().Clauses(clause.OnConflict{
		OnConstraint: "trans_branch_op_pkey",
		DoUpdates:    clause.AssignmentColumns(updates),
	}).Create(branches)
	return int(db.RowsAffected), db.Error
}

// LockGlobalSaveBranches saves branches in transaction
func (s *Store) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {
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

// MaySaveNewTrans creates branches or return error if conflict
func (s *Store) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	return dbGet().Transaction(func(db1 *gorm.DB) error {
		db := &dtmutil.DB{DB: db1}
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

// ChangeGlobalStatus changes global transaction status
func (s *Store) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	dbr := dbGet().Must().Model(global).Where("status=? and gid=?", old, global.Gid).Select(updates).Updates(global)
	if dbr.RowsAffected == 0 {
		dtmimp.E2P(storage.ErrNotFound)
	}
}

// TouchCronTime sets cron time
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64) {
	global.NextCronTime = dtmutil.GetNextTime(nextCronInterval)
	global.UpdateTime = dtmutil.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	dbGet().Must().Model(global).Where("status=? and gid=?", global.Status, global.Gid).
		Select([]string{"next_cron_time", "update_time", "next_cron_interval"}).Updates(global)
}

// LockOneGlobalTrans updates global transaction and return the latest.
func (s *Store) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	db := dbGet()
	getTime := func(second int) string {
		return map[string]string{
			"mysql":    fmt.Sprintf("date_add(now(), interval %d second)", second),
			"postgres": fmt.Sprintf("current_timestamp + interval '%d second'", second),
		}[conf.Store.Driver]
	}
	expire := int(expireIn / time.Second)
	whereTime := fmt.Sprintf("next_cron_time < %s", getTime(expire))
	owner := shortuuid.New()
	global := &storage.TransGlobalStore{}
	dbr := db.Must().Model(global).
		Where(whereTime + "and status in ('prepared', 'aborting', 'submitted')").
		Limit(1).
		Select([]string{"owner", "next_cron_time"}).
		Updates(&storage.TransGlobalStore{
			Owner:        owner,
			NextCronTime: dtmutil.GetNextTime(conf.RetryInterval),
		})
	if dbr.RowsAffected == 0 {
		return nil
	}
	db.Must().Where("owner=?", owner).First(global)
	return global
}

// SetDBConn sets db conn pool
func SetDBConn(db *gorm.DB) {
	sqldb, _ := db.DB()
	sqldb.SetMaxOpenConns(int(conf.Store.MaxOpenConns))
	sqldb.SetMaxIdleConns(int(conf.Store.MaxIdleConns))
	sqldb.SetConnMaxLifetime(time.Duration(conf.Store.ConnMaxLifeTime) * time.Minute)
}

func dbGet() *dtmutil.DB {
	return dtmutil.DbGet(conf.Store.GetDBConf(), SetDBConn)
}

func wrapError(err error) error {
	if err == gorm.ErrRecordNotFound {
		return storage.ErrNotFound
	}
	dtmimp.E2P(err)
	return err
}

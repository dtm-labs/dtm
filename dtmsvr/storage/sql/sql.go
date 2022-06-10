/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package sql

import (
	"fmt"
	"math"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/lithammer/shortuuid/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var conf = &config.Config

// Store implements storage.Store, and storage with db
type Store struct {
}

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

// FindTransGlobalStore finds GlobalTrans data by gid
func (s *Store) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	trans := &storage.TransGlobalStore{}
	dbr := dbGet().Model(trans).Where("gid=?", gid).First(trans)
	if dbr.Error == gorm.ErrRecordNotFound {
		return nil
	}
	dtmimp.E2P(dbr.Error)
	return trans
}

// ScanTransGlobalStores lists GlobalTrans data
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

// FindBranches finds Branch data by gid
func (s *Store) FindBranches(gid string) []storage.TransBranchStore {
	branches := []storage.TransBranchStore{}
	dbGet().Must().Where("gid=?", gid).Order("id asc").Find(&branches)
	return branches
}

// UpdateBranches update branches info
func (s *Store) UpdateBranches(branches []storage.TransBranchStore, updates []string) (int, error) {
	db := dbGet().Clauses(clause.OnConflict{
		OnConstraint: "trans_branch_op_pkey",
		DoUpdates:    clause.AssignmentColumns(updates),
	}).Create(branches)
	return int(db.RowsAffected), db.Error
}

// LockGlobalSaveBranches creates branches
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

// MaySaveNewTrans creates a new trans
func (s *Store) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	return dbGet().Transaction(func(db1 *gorm.DB) error {
		db := &dtmutil.DB{DB: db1}
		dbr := db.Must().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(global)
		if dbr.RowsAffected <= 0 { // not a new trans, return
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

// ChangeGlobalStatus changes global trans status
func (s *Store) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	dbr := dbGet().Must().Model(global).Where("status=? and gid=?", old, global.Gid).Select(updates).Updates(global)
	if dbr.RowsAffected == 0 {
		dtmimp.E2P(storage.ErrNotFound)
	}
}

// TouchCronTime updates cronTime
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64, nextCronTime *time.Time) {
	global.UpdateTime = dtmutil.GetNextTime(0)
	global.NextCronTime = nextCronTime
	global.NextCronInterval = nextCronInterval
	dbGet().Must().Model(global).Where("status=? and gid=?", global.Status, global.Gid).
		Select([]string{"next_cron_time", "update_time", "next_cron_interval"}).Updates(global)
}

// LockOneGlobalTrans finds GlobalTrans
func (s *Store) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	db := dbGet()
	owner := shortuuid.New()
	where := map[string]string{
		dtmimp.DBTypeMysql:    fmt.Sprintf(`next_cron_time < date_add(now(), interval %d second) and status in ('prepared', 'aborting', 'submitted') limit 1`, int(expireIn/time.Second)),
		dtmimp.DBTypePostgres: fmt.Sprintf(`id in (select id from trans_global where next_cron_time < current_timestamp + interval '%d second' and status in ('prepared', 'aborting', 'submitted') limit 1 )`, int(expireIn/time.Second)),
	}[conf.Store.Driver]

	sql := fmt.Sprintf(`UPDATE trans_global SET update_time='%s',next_cron_time='%s', owner='%s' WHERE %s`,
		getTimeStr(0),
		getTimeStr(conf.RetryInterval),
		owner,
		where)
	affected, err := dtmimp.DBExec(conf.Store.Driver, db.ToSQLDB(), sql)

	dtmimp.PanicIf(err != nil, err)
	if affected == 0 {
		return nil
	}
	global := &storage.TransGlobalStore{}
	db.Must().Where("owner=?", owner).First(global)
	return global
}

// ResetCronTime reset nextCronTime
// unfinished transactions need to be retried as soon as possible after business downtime is recovered
func (s *Store) ResetCronTime(after time.Duration, limit int64) (succeedCount int64, hasRemaining bool, err error) {
	where := map[string]string{
		dtmimp.DBTypeMysql:    fmt.Sprintf(`next_cron_time > date_add(now(), interval %d second) and status in ('prepared', 'aborting', 'submitted') limit %d`, int(after/time.Second), limit),
		dtmimp.DBTypePostgres: fmt.Sprintf(`id in (select id from trans_global where next_cron_time > current_timestamp + interval '%d second' and status in ('prepared', 'aborting', 'submitted') limit %d )`, int(after/time.Second), limit),
	}[conf.Store.Driver]

	sql := fmt.Sprintf(`UPDATE trans_global SET update_time='%s',next_cron_time='%s' WHERE %s`,
		getTimeStr(0),
		getTimeStr(0),
		where)
	affected, err := dtmimp.DBExec(conf.Store.Driver, dbGet().ToSQLDB(), sql)
	return affected, affected == limit, err
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

func getTimeStr(afterSecond int64) string {
	return dtmutil.GetNextTime(afterSecond).Format("2006-01-02 15:04:05")
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package boltdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
	bolt "go.etcd.io/bbolt"
)

// Store implements storage.Store, and storage with boltdb
type Store struct {
	boltDb *bolt.DB

	dataExpire    int64
	retryInterval int64
}

// NewStore will return the boltdb implement
// TODO: change to options
func NewStore(dataExpire int64, retryInterval int64) *Store {
	s := &Store{
		dataExpire:    dataExpire,
		retryInterval: retryInterval,
	}

	db, err := bolt.Open("./dtm.bolt", 0666, &bolt.Options{Timeout: 1 * time.Second})
	dtmimp.E2P(err)

	// NOTE: we must ensure all buckets is exists before we use it
	err = initializeBuckets(db)
	dtmimp.E2P(err)

	// TODO:
	//   1. refactor this code
	//   2. make cleanup run period, to avoid the file growup when server long-running
	err = cleanupExpiredData(
		time.Duration(dataExpire)*time.Second,
		db,
	)
	dtmimp.E2P(err)

	s.boltDb = db
	return s
}

func initializeBuckets(db *bolt.DB) error {
	return db.Update(func(t *bolt.Tx) error {
		for _, bucket := range allBuckets {
			_, err := t.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// cleanupExpiredData will clean the expired data in boltdb, the
//    expired time is configurable.
func cleanupExpiredData(expire time.Duration, db *bolt.DB) error {
	if expire <= 0 {
		return nil
	}

	lastKeepTime := time.Now().Add(-expire)
	return db.Update(func(t *bolt.Tx) error {
		globalBucket := t.Bucket(bucketGlobal)
		if globalBucket == nil {
			return nil
		}

		expiredGids := map[string]struct{}{}
		cursor := globalBucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			trans := storage.TransGlobalStore{}
			dtmimp.MustUnmarshal(v, &trans)

			transDoneTime := trans.FinishTime
			if transDoneTime == nil {
				transDoneTime = trans.RollbackTime
			}
			if transDoneTime != nil && lastKeepTime.After(*transDoneTime) {
				expiredGids[string(k)] = struct{}{}
			}
		}

		cleanupGlobalWithGids(t, expiredGids)
		cleanupBranchWithGids(t, expiredGids)
		cleanupIndexWithGids(t, expiredGids)
		return nil
	})
}

func cleanupGlobalWithGids(t *bolt.Tx, gids map[string]struct{}) {
	bucket := t.Bucket(bucketGlobal)
	if bucket == nil {
		return
	}

	logger.Debugf("Start to cleanup %d gids", len(gids))
	for gid := range gids {
		logger.Debugf("Start to delete gid: %s", gid)
		dtmimp.E2P(bucket.Delete([]byte(gid)))
	}
}

func cleanupBranchWithGids(t *bolt.Tx, gids map[string]struct{}) {
	bucket := t.Bucket(bucketBranches)
	if bucket == nil {
		return
	}

	// It's not safe if we delete the item when use cursor, for more detail see
	//    https://github.com/etcd-io/bbolt/issues/146
	branchKeys := []string{}
	for gid := range gids {
		cursor := bucket.Cursor()
		for k, v := cursor.Seek([]byte(gid)); k != nil; k, v = cursor.Next() {
			b := storage.TransBranchStore{}
			dtmimp.MustUnmarshal(v, &b)
			if b.Gid != gid {
				break
			}

			branchKeys = append(branchKeys, string(k))
		}
	}

	logger.Debugf("Start to cleanup %d branches", len(branchKeys))
	for _, key := range branchKeys {
		logger.Debugf("Start to delete branch: %s", key)
		dtmimp.E2P(bucket.Delete([]byte(key)))
	}
}

func cleanupIndexWithGids(t *bolt.Tx, gids map[string]struct{}) {
	bucket := t.Bucket(bucketIndex)
	if bucket == nil {
		return
	}

	indexKeys := []string{}
	cursor := bucket.Cursor()
	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
		ks := strings.Split(string(k), "-")
		if len(ks) != 2 {
			continue
		}

		if _, ok := gids[ks[1]]; ok {
			indexKeys = append(indexKeys, string(k))
		}
	}

	logger.Debugf("Start to cleanup %d indexes", len(indexKeys))
	for _, key := range indexKeys {
		logger.Debugf("Start to delete index: %s", key)
		dtmimp.E2P(bucket.Delete([]byte(key)))
	}
}

var bucketGlobal = []byte("global")
var bucketBranches = []byte("branches")
var bucketIndex = []byte("index")
var allBuckets = [][]byte{
	bucketBranches,
	bucketGlobal,
	bucketIndex,
}

func tGetGlobal(t *bolt.Tx, gid string) *storage.TransGlobalStore {
	trans := storage.TransGlobalStore{}
	bs := t.Bucket(bucketGlobal).Get([]byte(gid))
	if bs == nil {
		return nil
	}
	dtmimp.MustUnmarshal(bs, &trans)
	return &trans
}

func tGetBranches(t *bolt.Tx, gid string) []storage.TransBranchStore {
	branches := []storage.TransBranchStore{}
	cursor := t.Bucket(bucketBranches).Cursor()
	for k, v := cursor.Seek([]byte(gid)); k != nil; k, v = cursor.Next() {
		b := storage.TransBranchStore{}
		dtmimp.MustUnmarshal(v, &b)
		if b.Gid != gid {
			break
		}
		branches = append(branches, b)
	}
	return branches
}
func tPutGlobal(t *bolt.Tx, global *storage.TransGlobalStore) {
	bs := dtmimp.MustMarshal(global)
	err := t.Bucket(bucketGlobal).Put([]byte(global.Gid), bs)
	dtmimp.E2P(err)
}

func tPutBranches(t *bolt.Tx, branches []storage.TransBranchStore, start int64) {
	err := tPutBranches2(t, branches, start)
	dtmimp.E2P(err)
}

func tPutBranches2(t *bolt.Tx, branches []storage.TransBranchStore, start int64) error {
	if start == -1 {
		b0 := &branches[0]
		bs := tGetBranches(t, b0.Gid)
		for _, b := range bs {
			if b.BranchID == b0.BranchID && b.Op == b0.Op {
				return storage.ErrUniqueConflict
			}
		}
		start = int64(len(bs))
	}
	for i, b := range branches {
		k := b.Gid + fmt.Sprintf("%03d", i+int(start))
		v := dtmimp.MustMarshalString(b)
		err := t.Bucket(bucketBranches).Put([]byte(k), []byte(v))
		dtmimp.E2P(err)
	}
	return nil
}

func tDelIndex(t *bolt.Tx, unix int64, gid string) {
	k := fmt.Sprintf("%d-%s", unix, gid)
	err := t.Bucket(bucketIndex).Delete([]byte(k))
	dtmimp.E2P(err)
}

func tPutIndex(t *bolt.Tx, unix int64, gid string) {
	k := fmt.Sprintf("%d-%s", unix, gid)
	err := t.Bucket(bucketIndex).Put([]byte(k), []byte(gid))
	dtmimp.E2P(err)
}

// Ping execs ping cmd to boltdb
func (s *Store) Ping() error {
	return nil
}

// PopulateData populates data to boltdb
func (s *Store) PopulateData(skipDrop bool) {
	if !skipDrop {
		err := s.boltDb.Update(func(t *bolt.Tx) error {
			dtmimp.E2P(t.DeleteBucket(bucketIndex))
			dtmimp.E2P(t.DeleteBucket(bucketBranches))
			dtmimp.E2P(t.DeleteBucket(bucketGlobal))
			_, err := t.CreateBucket(bucketIndex)
			dtmimp.E2P(err)
			_, err = t.CreateBucket(bucketBranches)
			dtmimp.E2P(err)
			_, err = t.CreateBucket(bucketGlobal)
			dtmimp.E2P(err)

			return nil
		})
		dtmimp.E2P(err)
		logger.Infof("Reset all data for boltdb")
	}
}

// FindTransGlobalStore finds GlobalTrans data by gid
func (s *Store) FindTransGlobalStore(gid string) (trans *storage.TransGlobalStore) {
	err := s.boltDb.View(func(t *bolt.Tx) error {
		trans = tGetGlobal(t, gid)
		return nil
	})
	dtmimp.E2P(err)
	return
}

// ScanTransGlobalStores lists GlobalTrans data
func (s *Store) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
	globals := []storage.TransGlobalStore{}
	err := s.boltDb.View(func(t *bolt.Tx) error {
		cursor := t.Bucket(bucketGlobal).Cursor()
		for k, v := cursor.Seek([]byte(*position)); k != nil; k, v = cursor.Next() {
			if string(k) == *position {
				continue
			}
			g := storage.TransGlobalStore{}
			dtmimp.MustUnmarshal(v, &g)
			globals = append(globals, g)
			if len(globals) == int(limit) {
				break
			}
		}
		return nil
	})
	dtmimp.E2P(err)
	if len(globals) < int(limit) {
		*position = ""
	} else {
		*position = globals[len(globals)-1].Gid
	}
	return globals
}

// FindBranches finds Branch data by gid
func (s *Store) FindBranches(gid string) []storage.TransBranchStore {
	var branches []storage.TransBranchStore
	err := s.boltDb.View(func(t *bolt.Tx) error {
		branches = tGetBranches(t, gid)
		return nil
	})
	dtmimp.E2P(err)
	return branches
}

// UpdateBranches update branches info
func (s *Store) UpdateBranches(branches []storage.TransBranchStore, updates []string) (int, error) {
	return 0, nil // not implemented
}

// LockGlobalSaveBranches creates branches
func (s *Store) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {
	err := s.boltDb.Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, gid)
		if g == nil {
			return storage.ErrNotFound
		}
		if g.Status != status {
			return storage.ErrNotFound
		}
		return tPutBranches2(t, branches, int64(branchStart))
	})
	dtmimp.E2P(err)
}

// MaySaveNewTrans creates a new trans
func (s *Store) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	return s.boltDb.Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g != nil {
			return storage.ErrUniqueConflict
		}
		tPutGlobal(t, global)
		tPutIndex(t, global.NextCronTime.Unix(), global.Gid)
		tPutBranches(t, branches, 0)
		return nil
	})
}

// ChangeGlobalStatus changes global trans status
func (s *Store) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	err := s.boltDb.Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g == nil || g.Status != old {
			return storage.ErrNotFound
		}
		if finished {
			tDelIndex(t, g.NextCronTime.Unix(), g.Gid)
		}
		tPutGlobal(t, global)
		return nil
	})
	dtmimp.E2P(err)
}

// TouchCronTime updates cronTime
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64, nextCronTime *time.Time) {
	oldUnix := global.NextCronTime.Unix()
	global.UpdateTime = dtmutil.GetNextTime(0)
	global.NextCronTime = nextCronTime
	global.NextCronInterval = nextCronInterval
	err := s.boltDb.Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g == nil || g.Gid != global.Gid {
			return storage.ErrNotFound
		}
		tDelIndex(t, oldUnix, global.Gid)
		tPutGlobal(t, global)
		tPutIndex(t, global.NextCronTime.Unix(), global.Gid)
		return nil
	})
	dtmimp.E2P(err)
}

// LockOneGlobalTrans finds GlobalTrans
func (s *Store) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	var trans *storage.TransGlobalStore
	min := fmt.Sprintf("%d", time.Now().Add(expireIn).Unix())
	err := s.boltDb.Update(func(t *bolt.Tx) error {
		cursor := t.Bucket(bucketIndex).Cursor()
		toDelete := [][]byte{}
		for k, v := cursor.First(); k != nil && string(k) <= min && (trans == nil || trans.IsFinished()); k, v = cursor.Next() {
			trans = tGetGlobal(t, string(v))
			toDelete = append(toDelete, k)
		}
		for _, k := range toDelete {
			err := t.Bucket(bucketIndex).Delete(k)
			dtmimp.E2P(err)
		}
		if trans != nil && !trans.IsFinished() {
			next := time.Now().Add(time.Duration(s.retryInterval) * time.Second)
			trans.NextCronTime = &next
			tPutGlobal(t, trans)
			// this put should be after delete, because the data may be the same
			tPutIndex(t, next.Unix(), trans.Gid)
		}
		return nil
	})
	dtmimp.E2P(err)
	return trans
}

// ResetCronTime reset nextCronTime
// unfinished transactions need to be retried as soon as possible after business downtime is recovered
func (s *Store) ResetCronTime(after time.Duration, limit int64) (succeedCount int64, hasRemaining bool, err error) {
	next := time.Now()
	var trans *storage.TransGlobalStore
	min := fmt.Sprintf("%d", time.Now().Add(after).Unix())
	err = s.boltDb.Update(func(t *bolt.Tx) error {
		cursor := t.Bucket(bucketIndex).Cursor()
		succeedCount = 0
		for k, v := cursor.Seek([]byte(min)); k != nil && succeedCount <= limit; k, v = cursor.Next() {
			if succeedCount == limit {
				hasRemaining = true
				break
			}

			trans = tGetGlobal(t, string(v))
			err := t.Bucket(bucketIndex).Delete(k)
			dtmimp.E2P(err)

			trans.NextCronTime = &next
			tPutGlobal(t, trans)
			tPutIndex(t, next.Unix(), trans.Gid)
			succeedCount++
		}
		return nil
	})
	return
}

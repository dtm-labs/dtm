package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	bolt "go.etcd.io/bbolt"
	"gorm.io/gorm"
)

type BoltdbStore struct {
}

var boltDb *bolt.DB = nil
var boltOnce sync.Once

func boltGet() *bolt.DB {
	boltOnce.Do(func() {
		db, err := bolt.Open("./dtm.bolt", 0666, &bolt.Options{Timeout: 1 * time.Second})
		dtmimp.E2P(err)
		boltDb = db
	})
	return boltDb
}

var bucketGlobal = []byte("global")
var bucketBranches = []byte("branches")
var bucketIndex = []byte("index")

func tGetGlobal(t *bolt.Tx, gid string) *TransGlobalStore {
	trans := TransGlobalStore{}
	bs := t.Bucket(bucketGlobal).Get([]byte(gid))
	if bs == nil {
		return nil
	}
	dtmimp.MustUnmarshal(bs, &trans)
	return &trans
}

func tGetBranches(t *bolt.Tx, gid string) []TransBranchStore {
	branches := []TransBranchStore{}
	cursor := t.Bucket(bucketBranches).Cursor()
	for k, v := cursor.Seek([]byte(gid)); k != nil; k, v = cursor.Next() {
		b := TransBranchStore{}
		dtmimp.MustUnmarshal(v, &b)
		if b.Gid != gid {
			break
		}
		branches = append(branches, b)
	}
	return branches
}
func tPutGlobal(t *bolt.Tx, global *TransGlobalStore) {
	bs := dtmimp.MustMarshal(global)
	err := t.Bucket(bucketGlobal).Put([]byte(global.Gid), bs)
	dtmimp.E2P(err)
}

func tPutBranches(t *bolt.Tx, branches []TransBranchStore, start int64) {
	if start == -1 {
		bs := tGetBranches(t, branches[0].Gid)
		start = int64(len(bs))
	}
	for i, b := range branches {
		k := b.Gid + fmt.Sprintf("%03d", i+int(start))
		v := dtmimp.MustMarshalString(b)
		err := t.Bucket(bucketBranches).Put([]byte(k), []byte(v))
		dtmimp.E2P(err)
	}
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

func (s *BoltdbStore) PopulateData(skipDrop bool) {
	if !skipDrop {
		err := boltGet().Update(func(t *bolt.Tx) error {
			t.DeleteBucket(bucketIndex)
			t.DeleteBucket(bucketBranches)
			t.DeleteBucket(bucketGlobal)
			t.CreateBucket(bucketIndex)
			t.CreateBucket(bucketBranches)
			t.CreateBucket(bucketGlobal)
			return nil
		})
		dtmimp.E2P(err)
	}
}

func (s *BoltdbStore) FindTransGlobalStore(gid string) (trans *TransGlobalStore) {
	err := boltGet().View(func(t *bolt.Tx) error {
		trans = tGetGlobal(t, gid)
		return nil
	})
	dtmimp.E2P(err)
	return
}

func (s *BoltdbStore) ScanTransGlobalStores(position *string, limit int64) []TransGlobalStore {
	globals := []TransGlobalStore{}
	err := boltGet().View(func(t *bolt.Tx) error {
		cursor := t.Bucket(bucketGlobal).Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if string(k) == *position {
				continue
			}
			g := TransGlobalStore{}
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

func (s *BoltdbStore) FindBranches(gid string) []TransBranchStore {
	var branches []TransBranchStore = nil
	err := boltGet().View(func(t *bolt.Tx) error {
		branches = tGetBranches(t, gid)
		return nil
	})
	dtmimp.E2P(err)
	return branches
}

func (s *BoltdbStore) UpdateBranchesSql(branches []TransBranchStore, updates []string) *gorm.DB {
	return nil // not implemented
}

func (s *BoltdbStore) LockGlobalSaveBranches(gid string, status string, branches []TransBranchStore, branchStart int) {
	err := boltGet().Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, gid)
		if g == nil {
			return ErrNotFound
		}
		if g.Status != status {
			return ErrNotFound
		}
		tPutBranches(t, branches, int64(branchStart))
		return nil
	})
	dtmimp.E2P(err)
}

func (s *BoltdbStore) MaySaveNewTrans(global *TransGlobalStore, branches []TransBranchStore) error {
	return boltGet().Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g != nil {
			return ErrUniqueConflict
		}
		tPutGlobal(t, global)
		tPutIndex(t, global.NextCronTime.Unix(), global.Gid)
		tPutBranches(t, branches, 0)
		return nil
	})
}

func (s *BoltdbStore) ChangeGlobalStatus(global *TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	err := boltGet().Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g == nil || g.Status != old {
			return ErrNotFound
		}
		if finished {
			tDelIndex(t, g.NextCronTime.Unix(), g.Gid)
		}
		tPutGlobal(t, global)
		return nil
	})
	dtmimp.E2P(err)
}

func (s *BoltdbStore) TouchCronTime(global *TransGlobalStore, nextCronInterval int64) {
	oldUnix := global.NextCronTime.Unix()
	global.NextCronTime = common.GetNextTime(nextCronInterval)
	global.UpdateTime = common.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	err := boltGet().Update(func(t *bolt.Tx) error {
		g := tGetGlobal(t, global.Gid)
		if g == nil || g.Gid != global.Gid {
			return ErrNotFound
		}
		tDelIndex(t, oldUnix, global.Gid)
		tPutGlobal(t, global)
		tPutIndex(t, global.NextCronTime.Unix(), global.Gid)
		return nil
	})
	dtmimp.E2P(err)
}

func (s *BoltdbStore) LockOneGlobalTrans(expireIn time.Duration) *TransGlobalStore {
	var trans *TransGlobalStore = nil
	min := fmt.Sprintf("%d", time.Now().Add(expireIn).Unix())
	next := time.Now().Add(time.Duration(config.RetryInterval) * time.Second)
	err := boltGet().Update(func(t *bolt.Tx) error {
		cursor := t.Bucket(bucketIndex).Cursor()
		k, v := cursor.First()
		if k == nil || string(k) > min {
			return ErrNotFound
		}
		trans = tGetGlobal(t, string(v))
		err := t.Bucket(bucketIndex).Delete(k)
		dtmimp.E2P(err)
		if trans == nil { // index exists, but global trans not exists, so retry to get next
			return ErrShouldRetry
		}
		trans.NextCronTime = &next
		tPutGlobal(t, trans)
		tPutIndex(t, next.Unix(), trans.Gid)
		return nil
	})
	if err == ErrNotFound {
		return nil
	}
	dtmimp.E2P(err)
	return trans
}

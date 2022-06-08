package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/stretchr/testify/assert"
)

func initTransGlobal(gid string) (*storage.TransGlobalStore, storage.Store) {
	next := time.Now().Add(10 * time.Second)
	return initTransGlobalByNextCronTime(gid, next)
}

func initTransGlobalByNextCronTime(gid string, next time.Time) (*storage.TransGlobalStore, storage.Store) {
	g := &storage.TransGlobalStore{Gid: gid, Status: "prepared", NextCronTime: &next}
	bs := []storage.TransBranchStore{
		{Gid: gid, BranchID: "01"},
	}
	s := registry.GetStore()
	err := s.MaySaveNewTrans(g, bs)
	dtmimp.E2P(err)
	return g, s
}

func TestStoreSave(t *testing.T) {
	gid := dtmimp.GetFuncName()
	bs := []storage.TransBranchStore{
		{Gid: gid, BranchID: "01"},
		{Gid: gid, BranchID: "02"},
	}
	g, s := initTransGlobal(gid)
	g2 := s.FindTransGlobalStore(gid)
	assert.NotNil(t, g2)
	assert.Equal(t, gid, g2.Gid)

	bs2 := s.FindBranches(gid)
	assert.Equal(t, len(bs2), int(1))
	assert.Equal(t, "01", bs2[0].BranchID)

	s.LockGlobalSaveBranches(gid, g.Status, []storage.TransBranchStore{bs[1]}, -1)
	bs3 := s.FindBranches(gid)
	assert.Equal(t, 2, len(bs3))
	assert.Equal(t, "02", bs3[1].BranchID)
	assert.Equal(t, "01", bs3[0].BranchID)

	err := dtmimp.CatchP(func() {
		s.LockGlobalSaveBranches(g.Gid, "submitted", []storage.TransBranchStore{bs[1]}, 1)
	})
	assert.Equal(t, storage.ErrNotFound, err)

	s.ChangeGlobalStatus(g, "succeed", []string{}, true)
}

func TestStoreChangeStatus(t *testing.T) {
	gid := dtmimp.GetFuncName()
	g, s := initTransGlobal(gid)
	g.Status = "no"
	err := dtmimp.CatchP(func() {
		s.ChangeGlobalStatus(g, "submitted", []string{}, false)
	})
	assert.Equal(t, storage.ErrNotFound, err)
	g.Status = "prepared"
	s.ChangeGlobalStatus(g, "submitted", []string{}, false)
	s.ChangeGlobalStatus(g, "succeed", []string{}, true)
}

func TestStoreLockTrans(t *testing.T) {
	// lock trans will only lock unfinished trans. ensure all other trans are finished
	gid := dtmimp.GetFuncName()
	g, s := initTransGlobal(gid)

	g2 := s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.NotNil(t, g2)
	assert.Equal(t, gid, g2.Gid)

	s.TouchCronTime(g, 3*conf.RetryInterval, dtmutil.GetNextTime(3*conf.RetryInterval))
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.Nil(t, g2)

	s.TouchCronTime(g, 1*conf.RetryInterval, dtmutil.GetNextTime(1*conf.RetryInterval))
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.NotNil(t, g2)
	assert.Equal(t, gid, g2.Gid)

	s.ChangeGlobalStatus(g, "succeed", []string{}, true)
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.Nil(t, g2)
}

func TestStoreResetCronTime(t *testing.T) {
	s := registry.GetStore()
	testStoreResetCronTime(t, dtmimp.GetFuncName(), func(timeout int64, limit int64) (int64, bool, error) {
		return s.ResetCronTime(time.Duration(timeout)*time.Second, limit)
	})
}

func testStoreResetCronTime(t *testing.T, funcName string, resetCronHandler func(expire int64, limit int64) (int64, bool, error)) {
	s := registry.GetStore()
	var afterSeconds, lockExpireIn, limit, i int64
	afterSeconds = 100
	lockExpireIn = 2
	limit = 10

	// Will be reset
	for i = 0; i < limit; i++ {
		gid := funcName + fmt.Sprintf("%d", i)
		_, _ = initTransGlobalByNextCronTime(gid, time.Now().Add(time.Duration(afterSeconds+10)*time.Second))
	}

	// Will not be reset
	gid := funcName + fmt.Sprintf("%d", 10)
	_, _ = initTransGlobalByNextCronTime(gid, time.Now().Add(time.Duration(afterSeconds-10)*time.Second))

	// Not Found
	g := s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.Nil(t, g)

	// Reset limit-1 count
	succeedCount, hasRemaining, err := resetCronHandler(afterSeconds, limit-1)
	assert.Equal(t, hasRemaining, true)
	assert.Equal(t, succeedCount, limit-1)
	assert.Nil(t, err)
	// Found limit-1 count
	for i = 0; i < limit-1; i++ {
		g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
		assert.NotNil(t, g)
		s.ChangeGlobalStatus(g, "succeed", []string{}, true)
	}

	// Not Found
	g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.Nil(t, g)

	// Reset 1 count
	succeedCount, hasRemaining, err = resetCronHandler(afterSeconds, limit)
	assert.Equal(t, hasRemaining, false)
	assert.Equal(t, succeedCount, int64(1))
	assert.Nil(t, err)
	// Found 1 count
	g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.NotNil(t, g)
	s.ChangeGlobalStatus(g, "succeed", []string{}, true)

	// Not Found
	g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.Nil(t, g)

	// reduce the resetTimeTimeout, Reset 1 count
	succeedCount, hasRemaining, err = resetCronHandler(afterSeconds-12, limit)
	assert.Equal(t, hasRemaining, false)
	assert.Equal(t, succeedCount, int64(1))
	assert.Nil(t, err)
	// Found 1 count
	g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.NotNil(t, g)
	s.ChangeGlobalStatus(g, "succeed", []string{}, true)

	// Not Found
	g = s.LockOneGlobalTrans(time.Duration(lockExpireIn) * time.Second)
	assert.Nil(t, g)

	// Not Found
	succeedCount, hasRemaining, err = resetCronHandler(afterSeconds-12, limit)
	assert.Equal(t, hasRemaining, false)
	assert.Equal(t, succeedCount, int64(0))
	assert.Nil(t, err)
}

func TestUpdateBranches(t *testing.T) {
	if !conf.Store.IsDB() {
		_, err := registry.GetStore().UpdateBranches(nil, nil)
		assert.Nil(t, err)
	}
}

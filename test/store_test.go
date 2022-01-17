package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
)

func initTransGlobal(gid string) (*storage.TransGlobalStore, storage.Store) {
	next := time.Now().Add(10 * time.Second)
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

	s.TouchCronTime(g, 3*conf.RetryInterval)
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.Nil(t, g2)

	s.TouchCronTime(g, 1*conf.RetryInterval)
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.NotNil(t, g2)
	assert.Equal(t, gid, g2.Gid)

	s.ChangeGlobalStatus(g, "succeed", []string{}, true)
	g2 = s.LockOneGlobalTrans(2 * time.Duration(conf.RetryInterval) * time.Second)
	assert.Nil(t, g2)
}

func TestUpdateBranches(t *testing.T) {
	if !conf.Store.IsDB() {
		_, err := registry.GetStore().UpdateBranches(nil, nil)
		assert.Nil(t, err)
	}
}

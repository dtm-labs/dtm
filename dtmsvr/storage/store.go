package storage

import (
	"errors"
	"time"
)

// ErrNotFound defines the query item is not found in storage implement.
var ErrNotFound = errors.New("storage: NotFound")

// ErrUniqueConflict defines the item is conflict with unique key in storage implement.
var ErrUniqueConflict = errors.New("storage: UniqueKeyConflict")

type Store interface {
	Ping() error
	PopulateData(skipDrop bool)
	FindTransGlobalStore(gid string) *TransGlobalStore
	ScanTransGlobalStores(position *string, limit int64) []TransGlobalStore
	FindBranches(gid string) []TransBranchStore
	UpdateBranches(branches []TransBranchStore, updates []string) (int, error)
	LockGlobalSaveBranches(gid string, status string, branches []TransBranchStore, branchStart int)
	MaySaveNewTrans(global *TransGlobalStore, branches []TransBranchStore) error
	ChangeGlobalStatus(global *TransGlobalStore, newStatus string, updates []string, finished bool)
	TouchCronTime(global *TransGlobalStore, nextCronInterval int64)
	LockOneGlobalTrans(expireIn time.Duration) *TransGlobalStore
}

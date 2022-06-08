/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package storage

import (
	"errors"
	"time"
)

// ErrNotFound defines the query item is not found in storage implement.
var ErrNotFound = errors.New("storage: NotFound")

// ErrUniqueConflict defines the item is conflict with unique key in storage implement.
var ErrUniqueConflict = errors.New("storage: UniqueKeyConflict")

// Store defines storage relevant interface
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
	TouchCronTime(global *TransGlobalStore, nextCronInterval int64, nextCronTime *time.Time)
	LockOneGlobalTrans(expireIn time.Duration) *TransGlobalStore
	ResetCronTime(after time.Duration, limit int64) (succeedCount int64, hasRemaining bool, err error)
}

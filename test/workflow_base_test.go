/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowBranchConflict(t *testing.T) {
	gid := dtmimp.GetFuncName()
	store := dtmsvr.GetStore()
	now := time.Now()
	g := &storage.TransGlobalStore{
		Gid:          gid,
		Status:       dtmcli.StatusPrepared,
		NextCronTime: &now,
	}
	err := store.MaySaveNewTrans(g, []storage.TransBranchStore{
		{
			BranchID: "00",
			Op:       dtmimp.OpAction,
		},
	})
	assert.Nil(t, err)
	err = dtmimp.CatchP(func() {
		store.LockGlobalSaveBranches(gid, dtmcli.StatusPrepared, []storage.TransBranchStore{
			{BranchID: "00", Op: dtmimp.OpAction},
		}, -1)
	})
	assert.Error(t, err)
	store.ChangeGlobalStatus(g, StatusSucceed, []string{}, true)
}

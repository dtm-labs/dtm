/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmsvr/storage/registry"
)

type branchStatus struct {
	id         uint64
	status     string
	finishTime *time.Time
}

var p2e = dtmimp.P2E
var e2p = dtmimp.E2P

var config = &common.Config

func GetStore() storage.Store {
	return registry.GetStore()
}

// TransProcessedTestChan only for test usage. when transaction processed once, write gid to this chan
var TransProcessedTestChan chan string = nil

// GenGid generate gid, use uuid
func GenGid() string {
	return uuid.NewString()
}

// GetTransGlobal construct trans from db
func GetTransGlobal(gid string) *TransGlobal {
	trans := GetStore().FindTransGlobalStore(gid)
	dtmimp.PanicIf(trans == nil, fmt.Errorf("no TransGlobal with gid: %s found", gid))
	return &TransGlobal{TransGlobalStore: *trans}
}

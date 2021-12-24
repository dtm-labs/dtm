/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
)

func init() {
	addSample("saga", func() string {
		logger.Debugf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmHttpServer, dtmcli.MustGenGid(DtmHttpServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		logger.Debugf("saga busi trans submit")
		err := saga.Submit()
		logger.Debugf("result gid is: %s", saga.Gid)
		logger.FatalIfError(err)
		return saga.Gid
	})
	addSample("saga_wait", func() string {
		logger.Debugf("a saga busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmHttpServer, dtmcli.MustGenGid(DtmHttpServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req)
		saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
		err := saga.Submit()
		logger.Debugf("result gid is: %s", saga.Gid)
		logger.FatalIfError(err)
		return saga.Gid
	})
	addSample("concurrent_saga", func() string {
		logger.Debugf("a concurrent saga busi transaction begin")
		req := &TransReq{Amount: 30}
		csaga := dtmcli.NewSaga(DtmHttpServer, dtmcli.MustGenGid(DtmHttpServer)).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransOut", Busi+"/TransOutRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req).
			Add(Busi+"/TransIn", Busi+"/TransInRevert", req).
			EnableConcurrent().
			AddBranchOrder(2, []int{0, 1}).
			AddBranchOrder(3, []int{0, 1})
		logger.Debugf("concurrent saga busi trans submit")
		err := csaga.Submit()
		logger.Debugf("result gid is: %s", csaga.Gid)
		logger.FatalIfError(err)
		return csaga.Gid
	})
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func init() {
	setupFuncs["SagaBarrierSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/SagaBTransIn", common.WrapHandler(sagaBarrierTransIn))
		app.POST(BusiAPI+"/SagaBTransInCompensate", common.WrapHandler(sagaBarrierTransInCompensate))
		app.POST(BusiAPI+"/SagaBTransOut", common.WrapHandler(sagaBarrierTransOut))
		app.POST(BusiAPI+"/SagaBTransOutCompensate", common.WrapHandler(sagaBarrierTransOutCompensate))
	}
	addSample("saga_barrier", func() string {
		dtmimp.Logf("a busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmHttpServer, dtmcli.MustGenGid(DtmHttpServer)).
			Add(Busi+"/SagaBTransOut", Busi+"/SagaBTransOutCompensate", req).
			Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
		dtmimp.Logf("busi trans submit")
		err := saga.Submit()
		dtmimp.FatalIfError(err)
		return saga.Gid
	})
}

func sagaBarrierAdjustBalance(db dtmcli.DB, uid int, amount int) error {
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err

}

func sagaBarrierTransIn(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransInResult != "" {
		return req.TransInResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaBarrierAdjustBalance(tx, 1, req.Amount)
	})
}

func sagaBarrierTransInCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaBarrierAdjustBalance(tx, 1, -reqFrom(c).Amount)
	})
}

func sagaBarrierTransOut(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	if req.TransOutResult != "" {
		return req.TransOutResult, nil
	}
	barrier := MustBarrierFromGin(c)
	return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaBarrierAdjustBalance(tx, 2, -req.Amount)
	})
}

func sagaBarrierTransOutCompensate(c *gin.Context) (interface{}, error) {
	barrier := MustBarrierFromGin(c)
	return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaBarrierAdjustBalance(tx, 2, reqFrom(c).Amount)
	})
}

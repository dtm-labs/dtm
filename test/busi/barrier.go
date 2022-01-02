/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package busi

import (
	"context"
	"database/sql"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	setupFuncs["TccBarrierSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/SagaBTransIn", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			barrier := MustBarrierFromGin(c)
			return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
				return sagaAdjustBalance(tx, transInUID, reqFrom(c).Amount, reqFrom(c).TransInResult)
			})
		}))
		app.POST(BusiAPI+"/SagaBTransInCompensate", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			barrier := MustBarrierFromGin(c)
			return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
				return sagaAdjustBalance(tx, transInUID, -reqFrom(c).Amount, "")
			})
		}))
		app.POST(BusiAPI+"/SagaBTransOut", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			barrier := MustBarrierFromGin(c)
			return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
				return sagaAdjustBalance(tx, transOutUID, -reqFrom(c).Amount, reqFrom(c).TransOutResult)
			})
		}))
		app.POST(BusiAPI+"/SagaBTransOutCompensate", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			barrier := MustBarrierFromGin(c)
			return dtmcli.MapSuccess, barrier.Call(txGet(), func(tx *sql.Tx) error {
				return sagaAdjustBalance(tx, transOutUID, reqFrom(c).Amount, "")
			})
		}))
		app.POST(BusiAPI+"/SagaBTransOutGorm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			req := reqFrom(c)
			barrier := MustBarrierFromGin(c)
			tx := dbGet().DB.Begin()
			return dtmcli.MapSuccess, barrier.Call(tx.Statement.ConnPool.(*sql.Tx), func(tx1 *sql.Tx) error {
				return tx.Exec("update dtm_busi.user_account set balance = balance + ? where user_id = ?", -req.Amount, transOutUID).Error
			})
		}))

		app.POST(BusiAPI+"/TccBTransInTry", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			req := reqFrom(c) // 去重构一下，改成可以重复使用的输入
			if req.TransInResult != "" {
				return req.TransInResult, nil
			}
			return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
				return tccAdjustTrading(tx, transInUID, req.Amount)
			})
		}))
		app.POST(BusiAPI+"/TccBTransInConfirm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
				return tccAdjustBalance(tx, transInUID, reqFrom(c).Amount)
			})
		}))
		app.POST(BusiAPI+"/TccBTransInCancel", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
				return tccAdjustTrading(tx, transInUID, -reqFrom(c).Amount)
			})
		}))
		app.POST(BusiAPI+"/TccBTransOutTry", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			req := reqFrom(c)
			if req.TransOutResult != "" {
				return req.TransOutResult, nil
			}
			return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
				return tccAdjustTrading(tx, transOutUID, -req.Amount)
			})
		}))
		app.POST(BusiAPI+"/TccBTransOutConfirm", dtmutil.WrapHandler(func(c *gin.Context) (interface{}, error) {
			return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
				return tccAdjustBalance(tx, transOutUID, -reqFrom(c).Amount)
			})
		}))
		app.POST(BusiAPI+"/TccBTransOutCancel", dtmutil.WrapHandler(TccBarrierTransOutCancel))
	}
}

// TccBarrierTransOutCancel will be use in test
func TccBarrierTransOutCancel(c *gin.Context) (interface{}, error) {
	return dtmcli.MapSuccess, MustBarrierFromGin(c).Call(txGet(), func(tx *sql.Tx) error {
		return tccAdjustTrading(tx, transOutUID, reqFrom(c).Amount)
	})
}

func (s *busiServer) TransInBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaGrpcAdjustBalance(tx, transInUID, in.Amount, in.TransInResult)
	})
}

func (s *busiServer) TransOutBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaGrpcAdjustBalance(tx, transOutUID, -in.Amount, in.TransOutResult)
	})
}

func (s *busiServer) TransInRevertBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaGrpcAdjustBalance(tx, transInUID, -in.Amount, "")
	})
}

func (s *busiServer) TransOutRevertBSaga(ctx context.Context, in *BusiReq) (*emptypb.Empty, error) {
	barrier := MustBarrierFromGrpc(ctx)
	return &emptypb.Empty{}, barrier.Call(txGet(), func(tx *sql.Tx) error {
		return sagaGrpcAdjustBalance(tx, transOutUID, in.Amount, "")
	})
}

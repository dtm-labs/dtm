/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func init() {
	setupFuncs["TccSetupSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/TransInTccParent", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			tcc, err := dtmcli.TccFromQuery(c.Request.URL.Query())
			logger.FatalIfError(err)
			logger.Debugf("TransInTccParent ")
			return tcc.CallBranch(&TransReq{Amount: reqFrom(c).Amount}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		}))
	}
	addSample("tcc_nested", func() string {
		gid := dtmcli.MustGenGid(DtmHttpServer)
		err := dtmcli.TccGlobalTransaction(DtmHttpServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInTccParent", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		})
		logger.FatalIfError(err)
		return gid
	})
	addSample("tcc", func() string {
		logger.Debugf("tcc simple transaction begin")
		gid := dtmcli.MustGenGid(DtmHttpServer)
		err := dtmcli.TccGlobalTransaction(DtmHttpServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOut", Busi+"/TransOutConfirm", Busi+"/TransOutRevert")
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&TransReq{Amount: 30}, Busi+"/TransIn", Busi+"/TransInConfirm", Busi+"/TransInRevert")
		})
		logger.FatalIfError(err)
		return gid
	})
}

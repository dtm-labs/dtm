/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/go-resty/resty/v2"
)

func init() {
	addSample("xa_gorm", func() string {
		gid := dtmcli.MustGenGid(DtmHttpServer)
		err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOutXaGorm")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInXa")
		})
		logger.FatalIfError(err)
		return gid
	})

}

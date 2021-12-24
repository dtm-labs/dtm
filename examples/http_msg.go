/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmcli/logger"
)

func init() {
	addSample("msg", func() string {
		dtmimp.Logf("a busi transaction begin")
		req := &TransReq{Amount: 30}
		msg := dtmcli.NewMsg(DtmHttpServer, dtmcli.MustGenGid(DtmHttpServer)).
			Add(Busi+"/TransOut", req).
			Add(Busi+"/TransIn", req)
		err := msg.Prepare(Busi + "/query")
		logger.FatalIfError(err)
		dtmimp.Logf("busi trans submit")
		err = msg.Submit()
		logger.FatalIfError(err)
		return msg.Gid
	})
}

/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	context "context"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtmdriver"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// MustGenGid must gen a gid from grpcServer
func MustGenGid(grpcServer string) string {
	dc := dtmgimp.MustGetDtmClient(grpcServer)
	r, err := dc.NewGid(context.Background(), &emptypb.Empty{})
	dtmimp.E2P(err)
	return r.Gid
}

// SetCurrentDBType set the current db type
func SetCurrentDBType(dbType string) {
	dtmcli.SetCurrentDBType(dbType)
}

// GetCurrentDBType set the current db type
func GetCurrentDBType() string {
	return dtmcli.GetCurrentDBType()
}

func UseDriver(driverName string) error {
	return dtmdriver.Use(driverName)
}

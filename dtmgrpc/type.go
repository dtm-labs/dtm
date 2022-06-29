/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmgrpc

import (
	context "context"
	"errors"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"github.com/dtm-labs/dtmdriver"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// DtmError2GrpcError translate dtm error to grpc error
func DtmError2GrpcError(res interface{}) error {
	e, ok := res.(error)
	if ok && errors.Is(e, dtmimp.ErrFailure) {
		return status.New(codes.Aborted, e.Error()).Err()
	} else if ok && errors.Is(e, dtmimp.ErrOngoing) {
		return status.New(codes.FailedPrecondition, e.Error()).Err()
	}
	return e
}

// GrpcError2DtmError translate grpc error to dtm error
func GrpcError2DtmError(err error) error {
	st, _ := status.FromError(err)
	if st != nil && st.Code() == codes.Aborted {
		// version lower then v1.10, will specify Ongoing in code Aborted
		if st.Message() == dtmcli.ResultOngoing {
			return dtmcli.ErrOngoing
		}
		return fmt.Errorf("%s. %w", st.Message(), dtmcli.ErrFailure)
	} else if st != nil && st.Code() == codes.FailedPrecondition {
		return fmt.Errorf("%s. %w", st.Message(), dtmcli.ErrOngoing)
	}
	return err
}

// MustGenGid must gen a gid from grpcServer
func MustGenGid(grpcServer string) string {
	dc := dtmgimp.MustGetDtmClient(grpcServer)
	r, err := dc.NewGid(context.Background(), &emptypb.Empty{})
	dtmimp.E2P(err)
	return r.Gid
}

// UseDriver use the specified driver to handle grpc urls
func UseDriver(driverName string) error {
	return dtmdriver.Use(driverName)
}

// AddUnaryInterceptor adds grpc.UnaryClientInterceptor
func AddUnaryInterceptor(interceptor grpc.UnaryClientInterceptor) {
	dtmgimp.ClientInterceptors = append(dtmgimp.ClientInterceptors, interceptor)
}

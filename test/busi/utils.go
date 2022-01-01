package busi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dbGet() *dtmutil.DB {
	return dtmutil.DbGet(BusiConf)
}

func pdbGet() *sql.DB {
	db, err := dtmimp.PooledDB(BusiConf)
	logger.FatalIfError(err)
	return db
}

func txGet() *sql.Tx {
	db := pdbGet()
	tx, err := db.Begin()
	logger.FatalIfError(err)
	return tx
}

func resetXaData() {
	if BusiConf.Driver != "mysql" {
		return
	}

	db := dbGet()
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}

// MustBarrierFromGin 1
func MustBarrierFromGin(c *gin.Context) *dtmcli.BranchBarrier {
	ti, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
	logger.FatalIfError(err)
	return ti
}

// MustBarrierFromGrpc 1
func MustBarrierFromGrpc(ctx context.Context) *dtmcli.BranchBarrier {
	ti, err := dtmgrpc.BarrierFromGrpc(ctx)
	logger.FatalIfError(err)
	return ti
}

// SetGrpcHeaderForHeadersYes interceptor to set head for HeadersYes
func SetGrpcHeaderForHeadersYes(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if r, ok := req.(*dtmgpb.DtmRequest); ok && strings.HasSuffix(r.Gid, "HeadersYes") {
		logger.Debugf("writing test_header:test to ctx")
		md := metadata.New(map[string]string{"test_header": "test"})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// SetHttpHeaderForHeadersYes interceptor to set head for HeadersYes
func SetHttpHeaderForHeadersYes(c *resty.Client, r *resty.Request) error {
	if b, ok := r.Body.(*dtmcli.Saga); ok && strings.HasSuffix(b.Gid, "HeadersYes") {
		logger.Debugf("set test_header for url: %s", r.URL)
		r.SetHeader("test_header", "yes")
	}
	return nil
}

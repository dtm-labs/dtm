package dtmgrpc

import (
	context "context"
	"fmt"
	"strings"
	sync "sync"

	"github.com/yedf/dtm/dtmcli"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var clients sync.Map

// GetGrpcConn 1
func GetGrpcConn(grpcServer string) (conn *grpc.ClientConn, rerr error) {
	v, ok := clients.Load(grpcServer)
	if !ok {
		dtmcli.Logf("grpc client connecting %s", grpcServer)
		conn, rerr := grpc.Dial(grpcServer, grpc.WithInsecure(), grpc.WithUnaryInterceptor(GrpcClientLog))
		if rerr == nil {
			clients.Store(grpcServer, conn)
			v = conn
			dtmcli.Logf("grpc client inited for %s", grpcServer)
		}
	}
	return v.(*grpc.ClientConn), rerr
}

// MustGetGrpcConn 1
func MustGetGrpcConn(grpcServer string) *grpc.ClientConn {
	conn, err := GetGrpcConn(grpcServer)
	dtmcli.E2P(err)
	return conn
}

// MustGetDtmClient 1
func MustGetDtmClient(grpcServer string) DtmClient {
	return NewDtmClient(MustGetGrpcConn(grpcServer))
}

// MustGenGid 1
func MustGenGid(grpcServer string) string {
	dc := MustGetDtmClient(grpcServer)
	r, err := dc.NewGid(context.Background(), &emptypb.Empty{})
	dtmcli.E2P(err)
	return r.Gid
}

// GetServerAndMethod 将grpc的url分解为server和method
func GetServerAndMethod(grpcURL string) (string, string) {
	fs := strings.Split(grpcURL, "/")
	server := fs[0]
	method := "/" + strings.Join(fs[1:], "/")
	return server, method
}

// GrpcServerLog 打印grpc服务端的日志
func GrpcServerLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	dtmcli.Logf("grpc server handling: %s %v", info.FullMethod, req)
	m, err := handler(ctx, req)
	res := fmt.Sprintf("grpc server handled: %s %v result: %v err: %v", info.FullMethod, req, m, err)
	if err != nil {
		dtmcli.LogRedf("%s", res)
	} else {
		dtmcli.Logf("%s", res)
	}
	return m, err
}

// GrpcClientLog 打印grpc服务端的日志
func GrpcClientLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	dtmcli.Logf("grpc client calling: %s%s %v", cc.Target(), method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	res := fmt.Sprintf("grpc client called: %s%s %v result: %v err: %v", cc.Target(), method, req, reply, err)
	if err != nil {
		dtmcli.LogRedf("%s", res)
	} else {
		dtmcli.Logf("%s", res)
	}
	return err
}

// Result2Error 将通用的result转成grpc的error
func Result2Error(res interface{}, err error) error {
	e := dtmcli.CheckResult(res, err)
	if e == dtmcli.ErrFailure {
		return status.New(codes.Aborted, fmt.Sprintf("failure: res: %v, err: %s", res, e.Error())).Err()
	} else if e == dtmcli.ErrPending {
		return status.New(codes.Unavailable, fmt.Sprintf("failure: res: %v, err: %s", res, e.Error())).Err()
	}
	return e
}

package dtmpb

import (
	context "context"
	"strings"

	"github.com/yedf/dtm/dtmcli"
	grpc "google.golang.org/grpc"
)

var clients = map[string]*grpc.ClientConn{}

// GetGrpcConn 1
func GetGrpcConn(grpcServer string) (conn *grpc.ClientConn, rerr error) {
	if clients[grpcServer] == nil {
		conn, err := grpc.Dial(grpcServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithUnaryInterceptor(GrpcClientLog))
		if err == nil {
			clients[grpcServer] = conn
			dtmcli.Logf("dtm client inited for %s", grpcServer)
		}
	}
	conn = clients[grpcServer]
	return
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
	log := dtmcli.If(err != nil, dtmcli.LogRedf, dtmcli.Logf).(dtmcli.LogFunc)
	log("grpc server handled: %s %v result: %v err: %v", info.FullMethod, req, m, err)
	return m, err
}

// GrpcClientLog 打印grpc服务端的日志
func GrpcClientLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	dtmcli.Logf("grpc client calling: %s%s %v", cc.Target(), method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	log := dtmcli.If(err != nil, dtmcli.LogRedf, dtmcli.Logf).(dtmcli.LogFunc)
	log("grpc client called: %s%s %v result: %v err: %v", cc.Target(), method, req, reply, err)
	return err
}
